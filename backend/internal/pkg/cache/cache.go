package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/google/uuid"
	"github.com/slotmachine/backend/internal/config"
)

type NewCacheParams struct {
	Bus         EventBus
	Channel     string
	Config      *config.Config
	RedisClient RedisCloser
}

type RedisCloser interface {
	Close() error
}

type Cache struct {
	instanceID  string
	local       *ristretto.Cache[string, any]
	bus         EventBus
	channel     string
	Group       singleflight.Group
	config      *config.Config
	redisClient RedisCloser
}

type EventBus interface {
	Publish(channel string, payload any) error
	Subscribe(channel string, handler func(payload []byte))
}

type CacheMessage struct {
	SenderID string `json:"sender_id"`
	Type     string `json:"type"`
	Key      string `json:"key"`
	TTL      int64  `json:"ttl"`
}

func NewCache(p NewCacheParams) *Cache {
	if p.Channel == "" {
		panic("channel is required")
	}

	local, err := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters: 1e6,     // Number of keys
		MaxCost:     1 << 28, // ~256 MB
		BufferItems: 64,
	})

	if err != nil {
		panic(err)
	}

	c := &Cache{
		local:       local,
		bus:         p.Bus,
		channel:     p.Channel,
		instanceID:  uuid.New().String(),
		config:      p.Config,
		redisClient: p.RedisClient,
	}
	c.Subscribe()
	return c
}

func (c *Cache) GetWithSingleflight(ctx context.Context, key string, out any, fn func() (interface{}, error), ttls ...*time.Duration) (res any, err error) {
	if result, found := c.Get(ctx, key); found {
		return result, nil
	}

	val, err, _ := c.Group.Do(key, func() (interface{}, error) {
		if result, found := c.Get(ctx, key); found {
			return result, nil
		}

		newVal, err := fn()
		if err != nil {
			return nil, err
		}
		if len(ttls) > 0 && ttls[0] != nil {
			c.Set(ctx, key, newVal, *ttls[0])
		} else {
			c.Set(ctx, key, newVal, 0)
		}

		return newVal, nil
	})

	return val, err
}

func (c *Cache) Get(ctx context.Context, key string) (any, bool) {
	return c.local.Get(key)
}

func (c *Cache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if ttl == 0 {
		ttl = 30 * time.Minute
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	c.local.SetWithTTL(key, value, int64(len(data)), ttl)
	c.local.Wait()

	if c.bus != nil {
		msg := CacheMessage{Key: key, Type: "expired", SenderID: c.instanceID}

		publishData, _ := json.Marshal(msg)
		err := c.bus.Publish(c.channel, publishData)
		if err != nil {
			fmt.Printf("Failed to publish cache message for key: %s, err: %s", key, err.Error())
		}

		return err
	}

	return nil
}

func (c *Cache) Expire(ctx context.Context, key string) error {
	if c.bus == nil {
		return nil
	}
	val, found := c.local.Get(key)
	if !found {
		return nil
	}

	data, _ := json.Marshal(val)
	c.local.SetWithTTL(key, val, int64(len(data)), time.Millisecond)
	c.local.Wait()

	if c.bus != nil {
		msg := CacheMessage{Key: key, Type: "expired", SenderID: c.instanceID}
		publishData, _ := json.Marshal(msg)
		err := c.bus.Publish(c.channel, publishData)
		if err != nil {
			fmt.Printf("Failed to publish cache message for key: %s, err: %s", key, err.Error())
		}

		return err
	}

	return nil
}

func (c *Cache) Close() {
	// Close local cache
	c.local.Close()

	// Close Redis connection if available
	if c.redisClient != nil {
		if err := c.redisClient.Close(); err != nil {
			fmt.Printf("Failed to close Redis client: %s\n", err.Error())
		}
	}
}

func (c *Cache) Subscribe() {
	if c.bus == nil {
		return
	}

	c.bus.Subscribe(c.channel, func(payload []byte) {
		var msg CacheMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			return
		}

		if msg.SenderID != c.instanceID {
			if val, found := c.local.Get(msg.Key); found {
				switch msg.Type {
				case "expired":
					c.local.SetWithTTL(msg.Key, val, 1, time.Duration(msg.TTL)*time.Millisecond)
				}
			}
		}
	})
}
