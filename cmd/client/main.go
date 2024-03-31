package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"pow/internal/client"
	"pow/pkg/config"
	"pow/pkg/hashcash"
	"pow/pkg/protocol"
	"pow/pkg/protocol/pow"
)

func main() {
	// Config.
	cfg := Config{}
	err := config.LoadConfig("POW_CLIENT", &cfg)
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

	// Proto.
	proto := pow.NewPoW(logger.WithField("module", "pow"), cfg.Version, cfg.Target, h, 0)

	// Client.
	c := client.NewClient(logger.WithField("module", "client"), proto)

	//
	host, _, err := net.SplitHostPort(cfg.ServerAddr)
	if err != nil {
		panic(err)
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		panic(err)
	}
	logger.WithField("ips", ips).Debug()
	//

	doParams := client.DoParams{
		ServerAddr: cfg.ServerAddr,
		Network:    "tcp",
		Method:     protocol.SMGetQuote,
	}

	// Run client in cycle and exit.
	for i := 0; i < 3; i++ {
		ctxClient, cancelClient := context.WithTimeout(context.Background(), cfg.RequestTimeout)

		res, err := c.Do(ctxClient, doParams)
		if err != nil {
			fmt.Printf("Got error: '%s'\n", err.Error())
		} else {
			fmt.Println("GOT QUOTE: ", string(res))
		}

		cancelClient()
	}
}
