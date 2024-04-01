package client

import "context"

type hashcash interface {
	// FindNonce searches a nonce with bruteforce algorithm.
	// Can run multiple workers concurrently (concurrency factor specified on service creation),
	// reducing the time spent to find matched nonce.
	FindNonce(ctx context.Context, input []byte, target uint) ([]byte, error)
}
