package hashcash

import (
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateHash(t *testing.T) {
	t.Parallel()

	service := NewHashcash(nil, 1)

	testCases := []struct {
		name        string
		inputBytes  []byte
		inputTarget uint
		expected    bool
	}{
		{
			name:        "too_big_target",
			inputBytes:  []byte{0x0, 0x0, 0x0}, // 3*8=24 bits.
			inputTarget: 25,
			expected:    false,
		},
		{
			name:        "target_equal_input_bits_hit",
			inputBytes:  []byte{0x0}, // 8 bits.
			inputTarget: 8,
			expected:    true,
		},
		{
			name:        "target_equal_input_bits_miss",
			inputBytes:  []byte{0x0, 0xFF}, // 8 zeros, 16 total.
			inputTarget: 16,
			expected:    false,
		},
		{
			name:        "common_case_hit",
			inputBytes:  []byte{0b00000111, 0b00100101}, // 5 zeros.
			inputTarget: 3,
			expected:    true,
		},
		{
			name:        "common_case_miss",
			inputBytes:  []byte{0b00000111, 0b00100101}, // 5 zeros.
			inputTarget: 6,
			expected:    false,
		},
		{
			name:        "empty_input_and_zero_target",
			inputBytes:  []byte{},
			inputTarget: 0,
			expected:    true,
		},
		{
			name:        "nil_input_and_zero_target",
			inputBytes:  nil,
			inputTarget: 0,
			expected:    true,
		},
		{
			name:        "zero_target",
			inputBytes:  []byte{0xFF, 0xFF},
			inputTarget: 0,
			expected:    true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual := service.ValidateHash(tc.inputBytes, tc.inputTarget)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestHash(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		hashFunc func() hash.Hash
		input    []byte
	}{
		{
			name:     "sha1",
			hashFunc: sha1.New,
			input:    []byte{234, 78, 15, 43, 159},
		},
		{
			name:     "sha256",
			hashFunc: sha256.New,
			input:    []byte{240},
		},
		{
			name:     "sha512",
			hashFunc: sha512.New,
			input:    []byte{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service := NewHashcash(tc.hashFunc, 1)
			actual := service.Hash(tc.input)

			hasher := tc.hashFunc()
			_, _ = hasher.Write(tc.input)
			expected := hasher.Sum(nil)

			assert.Equal(t, expected, actual)
		})
	}
}

func TestFindNonce(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		input          []byte
		target         uint
		expectedErrStr string
	}{
		{
			name:           "target_too_big",
			input:          []byte("random byte slice"),
			target:         257,
			expectedErrStr: "hash size 256 less then target 257: wrong value",
		},
		{
			name:           "empty_input_is_correct_value",
			input:          nil,
			target:         3,
			expectedErrStr: "",
		},
		{
			name:           "common_case",
			input:          []byte{35, 137, 200},
			target:         3,
			expectedErrStr: "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			service := NewHashcash(sha256.New, 1)
			actual, err := service.FindNonce(ctx, tc.input, tc.target)

			if tc.expectedErrStr != "" {
				assert.EqualError(t, err, tc.expectedErrStr)
				assert.Nil(t, actual)
			} else {
				require.NoError(t, err)
				assert.True(t, service.ValidateHash(actual, tc.target))
			}

			cancel()
		})
	}
}
