package redisx

import (
    "context"

    redis "github.com/go-redis/redis/v8"
)

func New(addr string) *redis.Client {
    return redis.NewClient(&redis.Options{Addr: addr})
}

func Ping(ctx context.Context, c *redis.Client) error {
    return c.Ping(ctx).Err()
}
