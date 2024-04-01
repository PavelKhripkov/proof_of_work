package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"pow/internal/client"
	"pow/pkg/config"
	"pow/pkg/hashcash"
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

	// Client.
	c := client.NewClient(logger.WithField("module", "client"), h)

	params := client.GetQuotesParams{
		ServerAddr:    cfg.ServerAddr,
		QuotesDesired: 5,
	}

	// Getting quotes.
	fmt.Println("Requesting quote from the server")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.RequestTimeout)

	engageClient(ctx, c, params)
	cancel()

	// Providing wrong challenge.
	fmt.Println("Providing wrong challenge to the server")
	ctx, cancel = context.WithTimeout(context.Background(), cfg.RequestTimeout)

	params.ProvideWrongChallenge = true
	engageClient(ctx, c, params)
	cancel()

	// Providing wrong nonce.
	fmt.Println("Providing wrong nonce to the server")
	ctx, cancel = context.WithTimeout(context.Background(), cfg.RequestTimeout)

	params.ProvideWrongChallenge = false
	params.ProvideWrongNonce = true
	engageClient(ctx, c, params)
	cancel()

	// Too few time to calculate nonce.
	fmt.Println("Limiting time to calculate nonce")
	ctx, cancel = context.WithTimeout(context.Background(), 1)

	params.ProvideWrongNonce = false
	engageClient(ctx, c, params)
	cancel()
}

func engageClient(ctx context.Context, c *client.Client, params client.GetQuotesParams) {
	res, err := c.GetQuotes(ctx, params)
	if err != nil {
		fmt.Println("Got error, exiting: ", err.Error())
	} else {
		for _, quote := range res {
			fmt.Println("A WISE QUOTE FROM THE SERVER: ", quote)
		}
	}
}
