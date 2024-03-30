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
	"pow/pkg/protocol"
	"testing"
	"time"
)

type mockWriterClientRequest struct {
	outError error
}

func (s *mockWriterClientRequest) Write(input []byte) (int, error) {
	if s.outError != nil {
		return 0, s.outError
	}

	if len(input) != clientRequestSize {
		return 0, ErrWrongValue
	}

	req := clientRequest{}
	err := req.Unmarshal(input)
	if err != nil {
		return 0, err
	}

	return clientRequestSize, nil
}

type mockWriterServerResponse struct {
	outError error
}

func (s *mockWriterServerResponse) Write(input []byte) (int, error) {
	if s.outError != nil {
		return 0, s.outError
	}

	resp := serverResponse{}
	err := resp.Unmarshal(input)
	if err != nil {
		return 0, ErrWrongValue
	}

	switch {
	case
		resp.Code != protocol.SRCOK,
		string(resp.Body) != "some payload":
		return 0, ErrWrongValue
	default:
		return len(input), nil
	}
}

func TestSendClientRequest(t *testing.T) {
	t.Parallel()

	l := logrus.NewEntry(logrus.New())
	hc := hashcase2.NewHashcash(sha1.New, 1)
	pow := NewPoW(l, 10, 5, hc, 0)

	errSomeError := errors.New("some error")

	testCases := []struct {
		name         string
		inputWriter  io.Writer
		inputLocalIP string
		inputMethod  protocol.SeverMethod
		expectedErr  error
	}{
		{
			name:         "empty_local_ip",
			inputWriter:  nil,
			inputLocalIP: "",
			inputMethod:  protocol.SMGetQuote,
			expectedErr:  ErrWrongValue,
		},
		{
			name:         "conn_problem",
			inputWriter:  &mockWriterClientRequest{errSomeError},
			inputLocalIP: "127.0.0.1",
			inputMethod:  protocol.SMGetQuote,
			expectedErr:  errSomeError,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			err := pow.SendClientRequest(ctx, tc.inputWriter, tc.inputLocalIP, tc.inputMethod)
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
				return
			}
			assert.NoError(t, err)
		})
	}

}

func TestReceiveServerResponse(t *testing.T) {
	t.Parallel()

	correctServerResponse := serverResponse{
		Code: protocol.SRCOK,
		Body: []byte("some response body"),
	}
	correctServerResponseBytes := correctServerResponse.Marshal()

	errSomeError := errors.New("some error")
	pow := NewPoW(nil, 3, 20, nil, 0)

	testCases := []struct {
		name         string
		inputReader  io.Reader
		expectedCode protocol.ServerResponseCode
		expectedBody []byte
		expectedErr  error
	}{
		{
			name: "error_from_reader",
			inputReader: &mockReader{
				data: []byte{45},
				err:  errSomeError,
			},
			expectedErr: errSomeError,
		},
		{
			name: "empty_response",
			inputReader: &mockReader{
				data: nil,
				err:  io.EOF,
			},
			expectedErr: ErrEmptyResponse,
		},
		{
			name: "invalid_server_response",
			inputReader: &mockReader{
				data: []byte{39},
				err:  io.EOF,
			},
			expectedErr: ErrUnknownProtocol,
		},
		{
			name: "correct_case",
			inputReader: &mockReader{
				data: correctServerResponseBytes,
				err:  io.EOF,
			},
			expectedCode: protocol.SRCOK,
			expectedBody: correctServerResponse.Body,
			expectedErr:  nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			code, body, err := pow.ReceiveServerResponse(tc.inputReader)
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
				assert.Equal(t, protocol.SRCUnknownError, code)
				assert.Nil(t, body)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectedCode, code)
			assert.Equal(t, tc.expectedBody, body)
		})
	}

}

func TestHandleClientRequest(t *testing.T) {
	t.Parallel()

	errSomeError := errors.New("some error")

	l := logrus.NewEntry(logrus.New())
	hc := hashcase2.NewHashcash(sha1.New, 1)
	pow := NewPoW(l, 3, 20, hc, time.Hour)
	pow2 := NewPoW(l, 0, 20, hc, time.Hour)

	// Preparing wrong header.
	headerWrongVer, err := pow2.prepareHeader(context.Background(), "10.32.75.40")
	require.NoError(t, err)

	requestWrongVer := clientRequest{
		Header: *headerWrongVer,
		Method: protocol.SMGetQuote,
	}

	requestBytesWrongVer := requestWrongVer.Marshal()

	// Preparing correct header.
	headerCorrect, err := pow.prepareHeader(context.Background(), "10.32.75.40")
	require.NoError(t, err)

	requestCorrect := clientRequest{
		Header: *headerCorrect,
		Method: protocol.SMGetQuote,
	}

	requestBytesCorrect := requestCorrect.Marshal()

	testCases := []struct {
		name          string
		inputReader   io.Reader
		inputClientIP string
		inputChecker  func(context.Context, string, []byte) (bool, error)
		expectedRes   protocol.SeverMethod
		expectedErr   error
	}{
		{
			name: "reader_error",
			inputReader: &mockReader{
				data: []byte{1, 2, 3},
				err:  ErrWrongValue,
			},
			inputClientIP: "10.32.75.40",
			inputChecker: func(ctx context.Context, s string, bytes []byte) (bool, error) {
				return true, nil
			},
			expectedErr: ErrWrongValue,
		},
		{
			name: "header_hash_validation_error",
			inputReader: &mockReader{
				data: []byte{1, 2, 3},
				err:  nil,
			},
			inputClientIP: "10.32.75.40",
			inputChecker: func(ctx context.Context, s string, bytes []byte) (bool, error) {
				return true, nil
			},
			expectedErr: ErrUnknownProtocol,
		},
		{
			name: "header_validation_error",
			inputReader: &mockReader{
				data: requestBytesWrongVer,
				err:  nil,
			},
			inputClientIP: "10.32.75.40",
			inputChecker: func(ctx context.Context, s string, bytes []byte) (bool, error) {
				return true, nil
			},
			expectedErr: ErrWrongVersion,
		},
		{
			name: "checker_returns_error",
			inputReader: &mockReader{
				data: requestBytesCorrect,
				err:  nil,
			},
			inputClientIP: "10.32.75.40",
			inputChecker: func(ctx context.Context, s string, bytes []byte) (bool, error) {
				return false, errSomeError
			},
			expectedErr: errSomeError,
		},
		{
			name: "checker_returns_false",
			inputReader: &mockReader{
				data: requestBytesCorrect,
				err:  nil,
			},
			inputClientIP: "10.32.75.40",
			inputChecker: func(ctx context.Context, s string, bytes []byte) (bool, error) {
				return false, nil
			},
			expectedErr: ErrHashAlreadyUsed,
		},
		{
			name: "positive_case",
			inputReader: &mockReader{
				data: requestBytesCorrect,
				err:  nil,
			},
			inputClientIP: "10.32.75.40",
			inputChecker: func(ctx context.Context, s string, bytes []byte) (bool, error) {
				return true, nil
			},
			expectedRes: protocol.SMGetQuote,
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			actual, err := pow.HandleClientRequest(ctx, tc.inputReader, tc.inputClientIP, tc.inputChecker)
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
				assert.Equal(t, protocol.SMNoOp, actual)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectedRes, actual)
		})
	}
}

func TestValidateHeader(t *testing.T) {
	t.Parallel()

	pow := NewPoW(nil, 3, 20, nil, time.Hour)

	testCases := []struct {
		name          string
		inputHeader   clientRequestHeader
		inputClientIP string
		expectedErr   error
	}{
		{
			name: "mismatch_header_and_client_ip",
			inputHeader: clientRequestHeader{
				Ver:      3,
				Bits:     20,
				Date:     time.Now(),
				Resource: net.ParseIP("127.0.0.1"),
				Counter:  0,
			},
			inputClientIP: "127.0.0.2",
			expectedErr:   ErrWrongClientID,
		},
		{
			name: "version_mismatch",
			inputHeader: clientRequestHeader{
				Ver:      4,
				Bits:     20,
				Date:     time.Now(),
				Resource: net.ParseIP("127.0.0.1"),
				Counter:  0,
			},
			inputClientIP: "127.0.0.1",
			expectedErr:   ErrWrongVersion,
		},
		{
			name: "client_specified_fewer_bits_than_server_requires",
			inputHeader: clientRequestHeader{
				Ver:      3,
				Bits:     19,
				Date:     time.Now(),
				Resource: net.ParseIP("127.0.0.1"),
				Counter:  0,
			},
			inputClientIP: "127.0.0.1",
			expectedErr:   ErrWrongTargetBits,
		},
		{
			name: "header_is_spoiled",
			inputHeader: clientRequestHeader{
				Ver:      3,
				Bits:     20,
				Date:     time.Now().Add(-2 * time.Hour),
				Resource: net.ParseIP("127.0.0.1"),
				Counter:  0,
			},
			inputClientIP: "127.0.0.1",
			expectedErr:   ErrInvalidHeaderTime,
		},
		{
			name: "header_is_too_far_in_future",
			inputHeader: clientRequestHeader{
				Ver:      3,
				Bits:     20,
				Date:     time.Now().Add(2 * time.Hour),
				Resource: net.ParseIP("127.0.0.1"),
				Counter:  0,
			},
			inputClientIP: "127.0.0.1",
			expectedErr:   ErrInvalidHeaderTime,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := pow.validateHeader(tc.inputHeader, tc.inputClientIP)
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestSendServerResponse(t *testing.T) {
	t.Parallel()

	errSomeError := errors.New("some error")

	pow := NewPoW(nil, 3, 20, nil, time.Hour)

	testCases := []struct {
		name         string
		inputWriter  io.Writer
		inputCode    protocol.ServerResponseCode
		inputPayload []byte
		expectedErr  error
	}{
		{
			name:         "error_on_write",
			inputWriter:  &mockWriterServerResponse{errSomeError},
			inputCode:    protocol.SRCOK,
			inputPayload: []byte("some payload"),
			expectedErr:  errSomeError,
		},
		{
			name:         "write_success",
			inputWriter:  &mockWriterServerResponse{},
			inputCode:    protocol.SRCOK,
			inputPayload: []byte("some payload"),
			expectedErr:  nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := pow.SendServerResponse(tc.inputWriter, tc.inputCode, tc.inputPayload)
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
				return
			}

			assert.NoError(t, err)
		})
	}
}
