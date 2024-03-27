package pow

import "github.com/pkg/errors"

var (
	ErrUnknownProtocol = errors.New("unknown protocol")
	ErrInvalidHeader   = errors.New("invalid header")
	ErrHashAlreadyUsed = errors.New("hash already used")
	ErrWrongValue      = errors.New("wrong value")
	ErrEmptyResponse   = errors.New("empty response")
)