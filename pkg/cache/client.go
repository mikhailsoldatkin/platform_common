package cache

import (
	"context"
	"time"
)

type RedisClient interface {
	HashSet(ctx context.Context, key string, values any) error
	Set(ctx context.Context, key string, value any) error
	HGetAll(ctx context.Context, key string) ([]any, error)
	Get(ctx context.Context, key string) (any, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Ping(ctx context.Context) error
}
