package pow

import "context"

type hashcash interface {
	// FindNonce searches a nonce with bruteforce algorithm.
	// Can run multiple workers concurrently (concurrency factor specified on service creation),
	// reducing the time spent to find matched nonce.
	FindNonce(ctx context.Context, input []byte, target uint) ([]byte, error)
	// Hash returns hash calculated from input using target hash func specified on service creation.
	Hash(input []byte) []byte
	// ValidateHash checks if first target bits of input are equal to zero.
	// Returns false in case input has fewer bits than specified by target.
	ValidateHash(input []byte, target uint) bool
}
