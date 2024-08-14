package redis

import (
	"context"
	"log"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/mikhailsoldatkin/platform_common/pkg/cache"
)

var _ cache.RedisClient = (*Client)(nil)

type handler func(ctx context.Context, conn redis.Conn) error

// Client represents a Redis client that interacts with a Redis database using a connection pool.
type Client struct {
	pool   *redis.Pool
	config Config
}

// NewClient creates a new instance of Client with the provided Redis connection pool and configuration.
func NewClient(pool *redis.Pool, config Config) *Client {
	return &Client{
		pool:   pool,
		config: config,
	}
}

// HashSet sets multiple fields in a Redis hash stored at key.
func (c *Client) HashSet(ctx context.Context, key string, values any) error {
	err := c.execute(ctx, func(ctx context.Context, conn redis.Conn) error {
		_, err := conn.Do("HSET", redis.Args{key}.AddFlat(values)...)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// Set sets the value of a key in Redis.
func (c *Client) Set(ctx context.Context, key string, value any) error {
	err := c.execute(ctx, func(ctx context.Context, conn redis.Conn) error {
		_, err := conn.Do("SET", redis.Args{key}.Add(value)...)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// HGetAll retrieves all fields and values of a Redis hash stored at key.
func (c *Client) HGetAll(ctx context.Context, key string) ([]any, error) {
	var values []any
	err := c.execute(ctx, func(ctx context.Context, conn redis.Conn) error {
		var errEx error
		values, errEx = redis.Values(conn.Do("HGETALL", key))
		if errEx != nil {
			return errEx
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return values, nil
}

// Get retrieves the value of a key in Redis.
func (c *Client) Get(ctx context.Context, key string) (any, error) {
	var value any
	err := c.execute(ctx, func(ctx context.Context, conn redis.Conn) error {
		var errEx error
		value, errEx = conn.Do("GET", key)
		if errEx != nil {
			return errEx
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return value, nil
}

// Expire sets a timeout on the specified key in Redis.
// The key will be automatically deleted after the specified expiration duration.
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	err := c.execute(ctx, func(ctx context.Context, conn redis.Conn) error {
		_, err := conn.Do("EXPIRE", key, int(expiration.Seconds()))
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// Delete removes the specified key from Redis.
func (c *Client) Delete(ctx context.Context, key string) error {
	err := c.execute(ctx, func(ctx context.Context, conn redis.Conn) error {
		_, err := conn.Do("DEL", key)
		return err
	})

	if err != nil {
		return err
	}

	return nil
}

// Ping checks the connection to the Redis server.
func (c *Client) Ping(ctx context.Context) error {
	err := c.execute(ctx, func(ctx context.Context, conn redis.Conn) error {
		_, err := conn.Do("PING")
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// execute is a helper method that wraps Redis operations with connection management and error handling.
func (c *Client) execute(ctx context.Context, handler handler) error {
	conn, err := c.getConnect(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			log.Printf("failed to close redis connection: %v\n", err)
		}
	}()

	err = handler(ctx, conn)
	if err != nil {
		return err
	}

	return nil
}

// getConnect retrieves a Redis connection from the connection pool, respecting the context's timeout.
func (c *Client) getConnect(ctx context.Context) (redis.Conn, error) {
	getConnTimeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(c.config.ConnTimeout))
	defer cancel()

	conn, err := c.pool.GetContext(getConnTimeoutCtx)
	if err != nil {
		log.Printf("failed to get Redis connection: %v\n", err)
		_ = conn.Close()
		return nil, err
	}

	return conn, nil
}
