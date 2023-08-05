package redismemo

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Compute is a lazy function that when evaluated returns a string value.
// Computes are used as a fallback if the cache does not hold the given key.
type Compute func() string

// Memo is a function that memoizes values in a cache. Memo returns the string
// identified by the given cache key. If the cache does not hold the key,
// it computes the value and stores it with the given expiration TTL.
type Memo func(ctx context.Context, key string, value Compute, exp time.Duration) (string, error)

// RedisMemo wraps a Redis client to provide a memoization function.
func RedisMemo(rdb *redis.Client) Memo {
	memo := &redisMemo{rdb: rdb}
	return memo.Get
}

type redisMemo struct {
	rdb *redis.Client
}

func (rm *redisMemo) Get(ctx context.Context, key string, value Compute, exp time.Duration) (string, error) {
	res, err := rm.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		val := value()
		if err := rm.insert(ctx, key, val, exp); err != nil {
			return "", err
		}
		return val, nil // todo: trust the obvious or do another roundtrip?
	}
	return res, err
}

func (rm *redisMemo) insert(ctx context.Context, key string, value string, exp time.Duration) error {
	return rm.rdb.Set(ctx, key, value, exp).Err()
}
