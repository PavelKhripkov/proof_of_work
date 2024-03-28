package main

import (
	"context"
	"crypto/sha256"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"os/signal"
	"pow/internal/server"
	"pow/internal/storage/redis_storage"
	"pow/pkg/config"
	"pow/pkg/hashcash"
	"pow/pkg/protocol/pow"
	"syscall"
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
	redisStorage := redis_storage.NewRedis(redisClient, cfg.HashExp)

	// Proto.
	proto := pow.NewPoW(logger.WithField("module", "pow"), cfg.Version, cfg.Target, h, cfg.HeaderTimeInterval)

	// Server.
	s := server.NewServer(logger.WithField("module", "server"), redisStorage, proto, cfg.ResponseTimeout)

	ip := net.ParseIP(cfg.ServerAddr)
	if ip == nil {
		logger.Info("Server address is not specified. Using default.")
	}

	if cfg.ServerPort == 0 {
		logger.Info("Server port is not specified. Port will be chosen automatically.")
	}

	addr := net.TCPAddr{
		IP:   ip,
		Port: cfg.ServerPort,
		Zone: "",
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
	err = s.Run(ctx, &addr)
	if err != nil {
		logger.WithError(err).Error("Server stopped.")
		cancel()
	}

	// Waiting server is done.
	s.Done()
}
