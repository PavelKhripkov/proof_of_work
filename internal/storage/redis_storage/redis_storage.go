package redis_storage

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"time"
)

// keyTemplate template of a key used to store client hashes.
const keyTemplate = "pow:{%s}:%x"

// redisStorage is a service that uses Redis to store and retrieve data.
type redisStorage struct {
	c               *redis.Client
	hashExpDuration time.Duration
}

// NewRedis returns a new Redis storage service.
func NewRedis(c *redis.Client, hashExpDuration time.Duration) *redisStorage {
	return &redisStorage{
		c:               c,
		hashExpDuration: hashExpDuration,
	}
}

// GetSetHashByClient returns boolean value showing whether provided hash is allowed for provided client.
// Same hash can be used by a client only once in specified period of time.
// To fulfill this requirement, service stores that hash and checks if the client already used it recently.
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
