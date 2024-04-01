package main

import (
	"context"
	"crypto/sha256"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"pow/internal/server"
	"pow/internal/storage/redis_storage"
	"pow/pkg/config"
	"pow/pkg/hashcash"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// Config.
	cfg := Config{}
	err := config.LoadConfig("POW_SERVER", &cfg)
	if err != nil {
		panic(err.Error())
	}

	// Logger.
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	lvl, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		lvl = logrus.InfoLevel
	}
	logger.SetLevel(lvl)

	// Hashcash.
	h := hashcash.NewHashcash(sha256.New, cfg.Concurrency)

	// Redis.
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		panic("couldn't parse redisURL")
	}
	redisClient := redis.NewClient(opt)
	if _, err = redisClient.Ping(ctx).Result(); err != nil {
		err = errors.Wrap(err, "couldn't connect to redis")
		panic(err)
	}

	redisStorage := redis_storage.NewRedis(redisClient)

	// Server.
	params := server.NewServerParams{
		Logger:          logger.WithField("module", "server"),
		Storage:         redisStorage,
		Hashcash:        h,
		RespTimeout:     cfg.ResponseTimeout,
		RefreshDeadline: true,
		ChallengeSize:   cfg.ChallengeSize,
		TargetBits:      cfg.TargetBits,
		ChallengeTTL:    cfg.ChallengeTTL,
	}
	s := server.NewServer(params)

	host, _, err := net.SplitHostPort(cfg.ServerAddr)
	if err != nil {
		err = errors.Wrap(err, "couldn't parse host:port from config")
		panic(err)
	}

	if host == "" {
		logger.Info("Server address is not specified. Using default.")
	}

	// Graceful shutdown on signal.
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan,
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)

		// Waiting for signal and finishing services.
		sig := <-sigChan
		logger.Infof("Received signal %s, exiting.", sig)
		cancel()
	}()

	// Starting server.
	logger.Info("Starting server.")
	err = s.Run(ctx, cfg.ServerAddr)
	if err != nil {
		logger.WithError(err).Error("Server stopped.")
		cancel()
	}

	// Waiting server is done.
	s.Done()
}
