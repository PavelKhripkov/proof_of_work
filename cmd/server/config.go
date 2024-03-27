package main

import "time"

type Config struct {
	Version         byte          `envconfig:"VERSION" default:"0"`
	Target          byte          `envconfig:"TARGET" default:"20"`
	Concurrency     uint          `envconfig:"CONCURRENCY" default:"1"`
	RedisURL        string        `envconfig:"REDIS_URL" default:"redis://localhost/"`
	ServerAddr      string        `envconfig:"SERVER_ADDR" default:""`
	ServerPort      int           `envconfig:"SERVER_PORT" default:"9000"`
	HashExp         time.Duration `envconfig:"HASH_EXP" default:"1h"` // 1 Hour
	LogLevel        string        `envconfig:"LOG_LEVEL" default:"debug"`
	ResponseTimeout time.Duration `envconfig:"RESPONSE_TIMEOUT" default:"1h"`
	HashRetention   time.Duration `envconfig:"RESPONSE_TIMEOUT" default:"1h"`
	// TODO set redis TTL
	// TODO set time window for client request
}
