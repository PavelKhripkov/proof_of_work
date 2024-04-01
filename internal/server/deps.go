package server

import (
	"context"
	"time"
)

type storage interface {
	// StoreChallenge stores a challenge into DB.
	StoreChallenge(ctx context.Context, challenge []byte, hashTTL time.Duration) (err error)
	// GetDelChallenge extracts a challenge from DB, so it can't be used multiple times.
	GetDelChallenge(ctx context.Context, challenge []byte) (bool, error)
}

type hashcash interface {
	// Hash returns hash calculated from input using target hash func specified on service creation.
	Hash(input []byte) []byte
	// ValidateHash checks if first target bits of input are equal to zero.
	// Returns false in case input has fewer bits than specified by target.
	ValidateHash(input []byte, target uint) bool
}
