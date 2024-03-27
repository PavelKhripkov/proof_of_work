package redis_storage

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"time"
)

const keyTemplate = "pow:{%s}:%x"

type redisStorage struct {
	c               *redis.Client
	hashExpDuration time.Duration
}

func NewRedis(c *redis.Client, hashExpDuration time.Duration) *redisStorage {
	return &redisStorage{
		c:               c,
		hashExpDuration: hashExpDuration,
	}
}

func (s *redisStorage) GetSetHashByClient(ctx context.Context, client string, hash []byte) (bool, error) {
	key := fmt.Sprintf(keyTemplate, client, hash)

	_, err := s.c.Get(ctx, key).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return false, errors.Wrap(err, "couldn't check key in redis")
	}

	if err == nil {
		return false, nil
	}

	s.c.Set(ctx, key, "", s.hashExpDuration)
	return true, nil
}
