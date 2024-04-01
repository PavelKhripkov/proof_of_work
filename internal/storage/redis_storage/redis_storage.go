package redis_storage

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

// keyTemplate template of a key used to store client hashes.
const keyTemplate = "pow:%x"

// redisStorage is a service that uses Redis to store and retrieve data.
type redisStorage struct {
	c *redis.Client
}

// NewRedis returns a new Redis storage service.
func NewRedis(c *redis.Client) *redisStorage {
	return &redisStorage{c: c}
}

// StoreChallenge stores a challenge into DB.
func (s *redisStorage) StoreChallenge(ctx context.Context, challenge []byte, hashTTL time.Duration) (err error) {
	key := fmt.Sprintf(keyTemplate, challenge)

	_, err = s.c.Set(ctx, key, "", hashTTL).Result()

	return
}

// GetDelChallenge extracts a challenge from DB, so it can't be used multiple times.
func (s *redisStorage) GetDelChallenge(ctx context.Context, challenge []byte) (bool, error) {
	key := fmt.Sprintf(keyTemplate, challenge)

	_, err := s.c.GetDel(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}

	if err != nil {
		return false, errors.Wrap(err, "couldn't check key in redis")
	}

	return true, nil
}
