package main

import (
	"context"
	"crypto/sha256"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"os/signal"
	"pow/internal/config"
	"pow/internal/protocol/pow"
	"pow/internal/server"
	"pow/internal/storage/redis_storage"
	"pow/pkg/hashcash"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	cfg := Config{}
	err := config.LoadConfig("POW_SERVER", &cfg)
	if err != nil {
		panic(err.Error())
	}

	version := cfg.Version
	target := cfg.Target
	concurr := cfg.Concurrency
	redisURL := cfg.RedisURL
	serverAddr := cfg.ServerAddr
	serverPost := cfg.ServerPort
	hashExp := cfg.HashExp
	logLevel := cfg.LogLevel
	responseTimeout := cfg.ResponseTimeout
	hashRetention := cfg.HashRetention

	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	lvl, err := logrus.ParseLevel(logLevel)
	if err != nil {
		lvl = logrus.InfoLevel
	}
	logger.SetLevel(lvl)

	h, err := hashcash.NewHashcash(sha256.New, concurr)
	if err != nil {
		panic(err)
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		panic("couldn't parse redisURL")
	}
	redisClient := redis.NewClient(opt)

	redisStorage := redis_storage.NewRedis(redisClient, hashExp)
	proto := pow.NewPoW(logger.WithField("module", "pow"), version, target, h)

	s := server.NewServer(logger.WithField("module", "server"), redisStorage, proto, responseTimeout, hashRetention)

	ip := net.ParseIP(serverAddr)
	if ip == nil {
		logger.Info("Server address is not specified. Using default.")
	}

	if serverPost == 0 {
		logger.Info("Server port is not specified. Port will be chosen automatically.")
	}

	addr := net.TCPAddr{
		IP:   ip,
		Port: serverPost,
		Zone: "",
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan,
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)

		// Waiting for signal and finishing services.
		sig := <-sigChan
		logger.Infof("received signal %s, exiting...", sig)

		cancel()
	}()

	err = s.Run(ctx, &addr)
	if err != nil {
		logger.WithError(err).Error("Server stopped with error.")
	}

	s.Done()

	// TODO stop gracefully.
}
