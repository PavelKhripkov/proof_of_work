package main

import "time"

// Config represents client config.
type Config struct {
	Version        byte          `envconfig:"VERSION" default:"0"`
	Target         byte          `envconfig:"TARGET" default:"20"`
	Concurrency    uint          `envconfig:"CONCURRENCY" default:"1"`
	ServerAddr     string        `envconfig:"SERVER_ADDR" default:"127.0.0.1:9000"`
	LogLevel       string        `envconfig:"LOG_LEVEL" default:"info"`
	RequestTimeout time.Duration `envconfig:"REQUEST_TIMEOUT" default:"1s"`
}
