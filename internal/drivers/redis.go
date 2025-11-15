package drivers

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	c *redis.Client
}

func NewRedis(addr string) *RedisClient {
	c := redis.NewClient(&redis.Options{Addr: addr})
	return &RedisClient{c}
}

func (r *RedisClient) Close() error { return r.c.Close() }

func (r *RedisClient) Publish(ctx context.Context, channel string, payload string) error {
	return r.c.Publish(ctx, channel, payload).Err()
}

func (r *RedisClient) Subscribe(ctx context.Context, channel string) *redis.PubSub {
	return r.c.Subscribe(ctx, channel)
}

func (r *RedisClient) SetPresence(ctx context.Context, userID string, ttl time.Duration) error {
	return r.c.Set(ctx, fmt.Sprintf("presence:%s", userID), "online", ttl).Err()
}

func (r *RedisClient) RemovePresence(ctx context.Context, userID string) error {
	return r.c.Del(ctx, fmt.Sprintf("presence:%s", userID)).Err()
}
