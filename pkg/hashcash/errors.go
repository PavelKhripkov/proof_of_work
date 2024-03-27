package hashcash

import "github.com/pkg/errors"

var (
	ErrWrongValue = errors.New("wrong value")
	ErrNoMatch    = errors.New("match not found")
)
