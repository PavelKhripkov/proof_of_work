package pow

import (
	"context"
	"crypto/sha1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net"
	hashcase2 "pow/pkg/hashcash"
	"testing"
	"time"
)

type mockReader struct {
	data []byte
	err  error
}

func (s *mockReader) Read(dest []byte) (int, error) {
	return copy(dest, s.data), s.err
}

func TestReadClientRequest(t *testing.T) {
	t.Parallel()

	errSomeError := errors.New("some error")
	pow := NewPoW(nil, 0, 0, nil, 0)

	testCases := []struct {
		name        string
		conn        io.Reader
		expectedRes []byte
		expectedErr error
	}{
		{
			name: "error_while_reading_data_too_short",
			conn: &mockReader{
				data: []byte{1, 2, 3},
				err:  errSomeError,
			},
			expectedRes: nil,
			expectedErr: errSomeError,
		},
		{
			name: "error_while_reading_data_exact_size",
			conn: &mockReader{
				data: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2}, // 36
				err:  errSomeError,
			},
			expectedRes: nil,
			expectedErr: errSomeError,
		},
		{
			name: "no_error_too_short_data",
			conn: &mockReader{
				data: []byte{1, 2, 3},
				err:  nil,
			},
			expectedRes: nil,
			expectedErr: ErrUnknownProtocol,
		},
		{
			name: "no_error_data_exact_size",
			conn: &mockReader{
				data: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2}, // 36
				err:  nil,
			},
			expectedRes: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2}, // 36
			expectedErr: nil,
		},
		{
			name: "no_error_data_too_long",
			conn: &mockReader{
				data: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 3, 4}, // 38
				err:  nil,
			},
			expectedRes: nil,
			expectedErr: ErrUnknownProtocol,
		},
		{
			name: "eof_data_exact_size",
			conn: &mockReader{
				data: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2}, // 36
				err:  io.EOF,
			},
			expectedRes: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2}, // 36
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual, err := pow.readClientRequest(tc.conn)
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
				assert.Nil(t, actual)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectedRes, actual)
		})
	}
}

func TestValidateHeaderHash(t *testing.T) {
	t.Parallel()

	hc := hashcase2.NewHashcash(sha1.New, 1)

	correctInput := []byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 210, 204, 204, 204, 204, 204, 204, 76}
	correctInputHash := hc.Hash(correctInput)
	valid := hc.ValidateHash(correctInputHash, 3)
	require.True(t, valid)

	incorrectInput := []byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 210, 204, 204, 204, 204, 204, 204, 77}
	incorrectInputHash := hc.Hash(incorrectInput)
	valid = hc.ValidateHash(incorrectInputHash, 3)
	require.False(t, valid)

	l := logrus.NewEntry(logrus.New())
	pow := NewPoW(l, 0, 3, hc, 0)

	testCases := []struct {
		name        string
		input       []byte
		expectedRes []byte
		expectedErr error
	}{
		{
			name:        "correct_input",
			input:       correctInput,
			expectedRes: correctInputHash,
			expectedErr: nil,
		},
		{
			name:        "incorrect_input",
			input:       incorrectInput,
			expectedRes: nil,
			expectedErr: ErrInvalidHeader,
		},
		{
			name:        "too_short_input",
			input:       []byte{1, 2, 3},
			expectedRes: nil,
			expectedErr: ErrUnknownProtocol,
		},
		{
			name:        "too_long_input",
			input:       append(correctInput, 48),
			expectedRes: nil,
			expectedErr: ErrUnknownProtocol,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual, err := pow.validateHeaderHash(tc.input)
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
				assert.Equal(t, tc.expectedRes, actual)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectedRes, actual)
		})
	}
}

func TestPrepareHeader(t *testing.T) {
	t.Parallel()

	hc := hashcase2.NewHashcash(sha1.New, 1)
	l := logrus.NewEntry(logrus.New())
	pow := NewPoW(l, 10, 5, hc, 0)

	testCases := []struct {
		name        string
		input       string
		expectedRes *clientRequestHeader
		expectedErr error
	}{
		{
			name:  "correct_case",
			input: "10.10.1.15",
			expectedRes: &clientRequestHeader{
				Ver:      10,
				Bits:     5,
				Date:     time.Now(),
				Resource: net.ParseIP("10.10.1.15"),
				Counter:  0, // Not known yet.
			},
			expectedErr: nil,
		},
		{
			name:        "empty_address",
			input:       "",
			expectedRes: nil,
			expectedErr: ErrWrongValue,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			actual, err := pow.prepareHeader(ctx, tc.input)
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
				assert.Nil(t, actual)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectedRes.Ver, actual.Ver)
			assert.Equal(t, tc.expectedRes.Bits, actual.Bits)
			assert.Equal(t, tc.expectedRes.Resource, actual.Resource)
			assert.InDelta(t, tc.expectedRes.Date.UnixNano(), actual.Date.UnixNano(), float64(time.Second))

			assert.True(t, hc.ValidateHash(hc.Hash(actual.Marshal()), uint(tc.expectedRes.Bits)))
		})
	}
}
