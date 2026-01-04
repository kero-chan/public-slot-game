package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisEventBus struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisEventBus(rdb *redis.Client) *RedisEventBus {
	if rdb == nil {
		return nil
	}

	return &RedisEventBus{
		client: rdb,
		ctx:    context.Background(),
	}
}

func (r *RedisEventBus) Publish(channel string, payload any) error {
	if r == nil || r.client == nil {
		return nil
	}

	return r.client.Publish(r.ctx, channel, payload).Err()
}

func (r *RedisEventBus) Subscribe(channel string, handler func(payload []byte)) {
	if r == nil || r.client == nil {
		return
	}

	go func() {
		sub := r.client.Subscribe(r.ctx, channel)
		ch := sub.Channel()

		for msg := range ch {
			handler([]byte(msg.Payload))
		}
	}()
}
