package main

import "time"

// Config represents server config.
type Config struct {
	Concurrency     uint          `envconfig:"CONCURRENCY" default:"1"`
	RedisURL        string        `envconfig:"REDIS_URL" default:"redis://localhost/"`
	ServerAddr      string        `envconfig:"SERVER_ADDR" default:":9000"`
	LogLevel        string        `envconfig:"LOG_LEVEL" default:"info"`
	ResponseTimeout time.Duration `envconfig:"RESPONSE_TIMEOUT" default:"3s"`
	ChallengeSize   byte          `envconfig:"CHALLENGE_SIZE" default:"16"`
	TargetBits      byte          `envconfig:"TARGET_BITS" default:"20"`
	ChallengeTTL    time.Duration `envconfig:"CHALLENGE_TTL" default:"10m"`
}
