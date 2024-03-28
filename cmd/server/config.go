package main

import "time"

// Config represents server config.
type Config struct {
	Version            byte          `envconfig:"VERSION" default:"0"`
	Target             byte          `envconfig:"TARGET" default:"20"`
	Concurrency        uint          `envconfig:"CONCURRENCY" default:"1"`
	RedisURL           string        `envconfig:"REDIS_URL" default:"redis://localhost/"`
	ServerAddr         string        `envconfig:"SERVER_ADDR" default:""`
	ServerPort         int           `envconfig:"SERVER_PORT" default:"9000"`
	LogLevel           string        `envconfig:"LOG_LEVEL" default:"info"`
	ResponseTimeout    time.Duration `envconfig:"RESPONSE_TIMEOUT" default:"3s"`
	HashExp            time.Duration `envconfig:"HASH_EXP" default:"1h"`
	HeaderTimeInterval time.Duration `envconfig:"HEADER_TIME_INTERVAL" default:"1h"`
}
