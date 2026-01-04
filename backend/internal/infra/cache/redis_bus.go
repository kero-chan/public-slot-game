package cache

import (
	"context"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/slotmachine/backend/internal/pkg/logger"
)

// RedisBus implements the EventBus interface using Redis pub/sub
type RedisBus struct {
	client     *redis.Client
	logger     *logger.Logger
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	subscribed bool
	mu         sync.Mutex
}

// NewRedisBus creates a new Redis event bus
func NewRedisBus(client *redis.Client, log *logger.Logger) *RedisBus {
	ctx, cancel := context.WithCancel(context.Background())
	return &RedisBus{
		client: client,
		logger: log,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Publish publishes a message to a Redis channel
func (r *RedisBus) Publish(channel string, payload any) error {
	if r.client == nil {
		// If Redis is not available, silently skip
		return nil
	}

	err := r.client.Publish(r.ctx, channel, payload).Err()
	if err != nil {
		r.logger.Error().
			Err(err).
			Str("channel", channel).
			Msg("Failed to publish message to Redis")
		return err
	}

	return nil
}

// Subscribe subscribes to a Redis channel and calls handler for each message
func (r *RedisBus) Subscribe(channel string, handler func(payload []byte)) {
	if r.client == nil {
		// If Redis is not available, silently skip
		return
	}

	r.mu.Lock()
	r.subscribed = true
	r.mu.Unlock()

	pubsub := r.client.Subscribe(r.ctx, channel)

	r.logger.Info().
		Str("channel", channel).
		Msg("Subscribed to Redis channel")

	// Start a goroutine to listen for messages with proper cancellation
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		defer pubsub.Close()

		ch := pubsub.Channel()
		for {
			select {
			case msg, ok := <-ch:
				if !ok {
					// Channel closed
					return
				}
				handler([]byte(msg.Payload))
			case <-r.ctx.Done():
				// Context cancelled, shutdown gracefully
				r.logger.Info().
					Str("channel", channel).
					Msg("Unsubscribing from Redis channel")
				return
			}
		}
	}()
}

// Close gracefully shuts down the Redis bus, stopping all subscriptions
func (r *RedisBus) Close() error {
	r.mu.Lock()
	subscribed := r.subscribed
	r.mu.Unlock()

	if !subscribed {
		return nil
	}

	r.logger.Info().Msg("Closing Redis bus, stopping subscriptions")

	// Cancel context to stop all goroutines
	r.cancel()

	// Wait for all goroutines to finish
	r.wg.Wait()

	r.logger.Info().Msg("Redis bus closed")
	return nil
}
