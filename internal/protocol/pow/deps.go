package pow

import "context"

type hashcash interface {
	FindNonce(ctx context.Context, input []byte, target uint) ([]byte, error)
	Hash(input []byte) []byte
	ValidateHash(input []byte, target uint) bool
}
