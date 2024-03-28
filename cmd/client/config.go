package main

import "time"

// Config represents client config.
type Config struct {
	Version        byte          `envconfig:"VERSION" default:"0"`
	Target         byte          `envconfig:"TARGET" default:"20"`
	Concurrency    uint          `envconfig:"CONCURRENCY" default:"1"`
	RemoteAddr     string        `envconfig:"REMOTE_ADDR" default:"localhost:9000"`
	LogLevel       string        `envconfig:"LOG_LEVEL" default:"debug"`
	RequestTimeout time.Duration `envconfig:"REQUEST_TIMEOUT" default:"1h"`
}
