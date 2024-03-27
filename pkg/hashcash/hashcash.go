package hashcash

import (
	"context"
	"encoding/binary"
	"golang.org/x/sync/errgroup"
	"hash"
	"math"
)

type hashcash struct {
	concurrency uint
	hashFunc    func() hash.Hash
}

func NewHashcash(hashFunc func() hash.Hash, concurrency uint) (*hashcash, error) {
	if concurrency == 0 {
		concurrency = 1
	}

	return &hashcash{
		concurrency: concurrency,
		hashFunc:    hashFunc,
	}, nil
}

func (s *hashcash) FindNonce(ctx context.Context, input []byte, target uint) ([]byte, error) {
	if uint(len(input)*8) < target {
		return nil, ErrWrongValue
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

func (s *hashcash) Hash(input []byte) []byte {
	hasher := s.hashFunc()
	_, _ = hasher.Write(input) // Never returns an error. https://pkg.go.dev/hash#Hash
	return hasher.Sum(nil)
}

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
