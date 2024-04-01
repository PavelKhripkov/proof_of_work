package main

import "time"

// Config represents client config.
type Config struct {
	Concurrency    uint          `envconfig:"CONCURRENCY" default:"1"`
	ServerAddr     string        `envconfig:"SERVER_ADDR" default:"127.0.0.1:9000"`
	LogLevel       string        `envconfig:"LOG_LEVEL" default:"info"`
	RequestTimeout time.Duration `envconfig:"REQUEST_TIMEOUT" default:"1s"`
}
