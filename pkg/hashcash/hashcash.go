package hashcash

import (
	"context"
	"encoding/binary"
	"hash"
	"math"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// hashcash represents hashcash service.
type hashcash struct {
	concurrency uint
	// hashFunc is a function that returns a new hasher. Client and server MUST specify the same one.
	hashFunc func() hash.Hash
}

// NewHashcash creates new hashcash instance.
//
// hashFunc argument MUST be the same for client and server.
func NewHashcash(hashFunc func() hash.Hash, concurrency uint) *hashcash {
	if concurrency == 0 {
		concurrency = 1
	}

	return &hashcash{
		concurrency: concurrency,
		hashFunc:    hashFunc,
	}
}

// FindNonce searches a nonce with bruteforce algorithm.
// Can run multiple workers concurrently (concurrency factor specified on service creation),
// reducing the time spent to find matched nonce.
func (s *hashcash) FindNonce(ctx context.Context, input []byte, target uint) ([]byte, error) {
	hashBitSize := uint(s.hashFunc().Size()) * 8
	if hashBitSize < target {
		return nil, errors.Wrapf(ErrWrongValue, "hash size %d less then target %d", hashBitSize, target)
	}

	batch := math.MaxUint64 / uint64(s.concurrency)
	remainder := math.MaxUint64 % s.concurrency
	start, end := uint64(0), batch

	ctxGParent, cancel := context.WithCancel(ctx)
	defer cancel()

	g, ctxG := errgroup.WithContext(ctxGParent)
	resChan := make(chan []byte)

	for i := uint(0); i < s.concurrency; i++ {
		if remainder != 0 {
			end++
			remainder--
		}

		currStart, currEnd := start, end
		start, end = end+1, end+1+batch

		g.Go(func() error {
			return s.findNonce(ctxG, s.hashFunc(), input, target, resChan, currStart, currEnd)
		})
	}

	go func() {
		_ = g.Wait()
		close(resChan)
	}()

	if res, ok := <-resChan; ok {
		cancel()
		_ = g.Wait()
		return res, nil
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return nil, ErrNoMatch
}

// findNonce is a worker of FindNonce that can be run in goroutine with specified nonce window.
func (s *hashcash) findNonce(ctx context.Context, hasher hash.Hash, input []byte, target uint, resChan chan<- []byte, startNonce, endNonce uint64) error {
	l := len(input)
	temp := make([]byte, l+8)
	copy(temp, input)
	binary.LittleEndian.PutUint64(temp[l:], startNonce)

	// helper increases nonce in-place.
	helper := func() {
		for i := l; i < len(temp); i++ {
			if temp[i] < 0xFF {
				temp[i]++
				return
			}
			temp[i] = 0x00
		}

		temp = append(temp, 0x01) // Overflow case. Can't be reached if algorithm works correctly.
		return
	}

	for curr := startNonce; curr <= endNonce; curr++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		hasher.Reset()
		_, _ = hasher.Write(temp) // Never returns an error. https://pkg.go.dev/hash#Hash

		if !s.ValidateHash(hasher.Sum(nil), target) {
			curr++
			helper()
			continue
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case resChan <- temp[len(input):]:
			return nil
		}
	}

	return nil
}

// Hash returns hash calculated from input using target hash func specified on service creation.
func (s *hashcash) Hash(input []byte) []byte {
	hasher := s.hashFunc()
	_, _ = hasher.Write(input) // Never returns an error. https://pkg.go.dev/hash#Hash
	return hasher.Sum(nil)
}

// ValidateHash checks if first target bits of input are equal to zero.
// Returns false in case input has fewer bits than specified by target.
func (s *hashcash) ValidateHash(input []byte, target uint) bool {
	if uint(len(input)*8) < target {
		return false
	}

	var i int

	for target >= 8 {
		if input[i] != 0 {
			return false
		}

		target -= 8
		i++
	}

	if target != 0 {
		if input[i]>>(8-target) != 0 {
			return false
		}
	}

	return true
}
