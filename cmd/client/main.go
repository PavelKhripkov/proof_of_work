package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"pow/internal/client"
	"pow/pkg/config"
	"pow/pkg/hashcash"
	"pow/pkg/protocol"
	"pow/pkg/protocol/pow"
)

func main() {
	cfg := Config{}
	err := config.LoadConfig("POW_CLIENT", &cfg)
	if err != nil {
		panic(err.Error())
	}

	version := cfg.Version
	target := cfg.Target
	concurr := cfg.Concurrency
	remoteAddr := cfg.RemoteAddr
	logLevel := cfg.LogLevel
	requestTimeout := cfg.RequestTimeout

	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	lvl, err := logrus.ParseLevel(logLevel)
	if err != nil {
		lvl = logrus.InfoLevel
	}
	logger.SetLevel(lvl)

	h := hashcash.NewHashcash(sha256.New, concurr)

	proto := pow.NewPoW(logger.WithField("module", "pow"), version, target, h)

	c := client.NewClient(logger.WithField("module", "client"), proto)

	doParams := client.DoParams{
		RemoteAddr: remoteAddr,
		Method:     protocol.SMGetQuote,
	}

	for i := 0; i < 3; i++ {
		ctxClient, cancelClient := context.WithTimeout(context.Background(), requestTimeout)

		res, err := c.Do(ctxClient, doParams)
		if err != nil {
			fmt.Printf("Got error: '%s'\n", err.Error())
		} else {
			fmt.Println(string(res))
		}

		cancelClient()
	}
}
