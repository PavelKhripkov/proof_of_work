package server

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pow/internal/models"
	hashcash2 "pow/pkg/hashcash"
)

var errSomeError = errors.New("some error")

type mockStorage struct {
	StoreChallengeResponse         error
	GetDelChallengeResponseAllowed bool
	GetDelChallengeResponseErr     error
}

func (s *mockStorage) StoreChallenge(ctx context.Context, challenge []byte, hashTTL time.Duration) (err error) {
	return s.StoreChallengeResponse
}

func (s *mockStorage) GetDelChallenge(ctx context.Context, challenge []byte) (bool, error) {
	return s.GetDelChallengeResponseAllowed, s.GetDelChallengeResponseErr
}

func TestGetChallengeHandler(t *testing.T) {
	t.Parallel()

	var target byte = 15

	strg := mockStorage{}
	params := NewServerParams{
		Logger:        nil,
		Hashcash:      nil,
		Storage:       &strg,
		RespTimeout:   0,
		ChallengeSize: 10,
		TargetBits:    target,
		ChallengeTTL:  0,
	}

	svr := NewServer(params)

	testCases := []struct {
		name            string
		storageResponse error
	}{
		{
			name:            "couldn't_store_challenge",
			storageResponse: errSomeError,
		},
		{
			name:            "positive_case",
			storageResponse: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			strg.StoreChallengeResponse = tc.storageResponse
			actual, err := svr.getChallengeHandler(nil)
			if tc.storageResponse != nil {
				assert.ErrorIs(t, err, tc.storageResponse)
				assert.Equal(t, models.SRCInternalError, actual.Code)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, models.SRCOK, actual.Code)
			assert.Equal(t, target, actual.Target)
			assert.Equal(t, 10, len(actual.Challenge))
		})
	}
}

func TestGetQuoteHandler(t *testing.T) {
	t.Parallel()

	var target byte = 15

	correctChallenge := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	correctNonce := []byte{157, 1, 1, 0, 0, 0, 0, 0}

	h := hashcash2.NewHashcash(sha256.New, 1)

	valid := h.ValidateHash(h.Hash(append(correctChallenge, correctNonce...)), uint(target))
	require.True(t, valid)

	strg := mockStorage{}
	params := NewServerParams{
		Logger:        nil,
		Hashcash:      h,
		Storage:       &strg,
		RespTimeout:   0,
		ChallengeSize: 10,
		TargetBits:    target,
		ChallengeTTL:  0,
	}

	svr := NewServer(params)

	testCases := []struct {
		name               string
		input              models.ClientRequest
		storageRespAllowed bool
		storageRespErr     error
		expectedCode       models.ServerResponseCode
		expectedErr        error
	}{
		{
			name: "check_challenge_error",
			input: models.ClientRequest{
				Method:    models.SMGetQuote,
				Challenge: correctChallenge,
				Nonce:     correctNonce,
			},
			storageRespAllowed: false,
			storageRespErr:     errSomeError,
			expectedCode:       models.SRCInternalError,
			expectedErr:        errSomeError,
		},
		{
			name: "check_challenge_not_passed",
			input: models.ClientRequest{
				Method:    models.SMGetQuote,
				Challenge: correctChallenge,
				Nonce:     correctNonce,
			},
			storageRespAllowed: false,
			storageRespErr:     nil,
			expectedCode:       models.SRCWrongChallenge,
			expectedErr:        nil,
		},
		{
			name: "positive_case",
			input: models.ClientRequest{
				Method:    models.SMGetQuote,
				Challenge: correctChallenge,
				Nonce:     correctNonce,
			},
			storageRespAllowed: true,
			storageRespErr:     nil,
			expectedCode:       models.SRCOK,
			expectedErr:        nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			strg.GetDelChallengeResponseAllowed = tc.storageRespAllowed
			strg.GetDelChallengeResponseErr = tc.storageRespErr

			actual, err := svr.getQuoteHandler(nil, tc.input)
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
				assert.Equal(t, tc.expectedCode, actual.Code)
				assert.Empty(t, actual.Body)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectedCode, actual.Code)

			if actual.Code != models.SRCOK {
				assert.Empty(t, actual.Body)
			} else {
				assert.NotEmpty(t, actual.Body)
			}
		})
	}
}

func TestCheckChallenge(t *testing.T) {
	t.Parallel()

	var target byte = 15

	correctChallenge := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	correctNonce := []byte{157, 1, 1, 0, 0, 0, 0, 0}

	incorrectChallenge := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	incorrectNonce := []byte{157, 1, 1, 0, 0, 0, 0, 1}

	h := hashcash2.NewHashcash(sha256.New, 1)

	valid := h.ValidateHash(h.Hash(append(correctChallenge, correctNonce...)), uint(target))
	require.True(t, valid)

	valid = h.ValidateHash(h.Hash(append(incorrectChallenge, incorrectNonce...)), uint(target))
	require.False(t, valid)

	strg := mockStorage{}
	params := NewServerParams{
		Logger:        nil,
		Hashcash:      h,
		Storage:       &strg,
		RespTimeout:   0,
		ChallengeSize: 10,
		TargetBits:    target,
		ChallengeTTL:  0,
	}

	svr := NewServer(params)

	testCases := []struct {
		name               string
		input              models.ClientRequest
		storageRespAllowed bool
		storageRespErr     error
		expectedCode       models.ServerResponseCode
		expectedErr        error
	}{
		{
			name:           "storage_error",
			storageRespErr: errSomeError,
			expectedCode:   models.SRCInternalError,
			expectedErr:    errSomeError,
		},
		{
			name:               "challenge_not_exists",
			storageRespAllowed: false,
			storageRespErr:     nil,
			expectedCode:       models.SRCWrongChallenge,
			expectedErr:        nil,
		},
		{
			name: "invalid_hash",
			input: models.ClientRequest{
				Method:    models.SMGetQuote,
				Challenge: incorrectChallenge,
				Nonce:     incorrectNonce,
			},
			storageRespAllowed: true,
			storageRespErr:     nil,
			expectedCode:       models.SRCWrongNonce,
			expectedErr:        nil,
		},
		{
			name: "positive_case",
			input: models.ClientRequest{
				Method:    models.SMGetQuote,
				Challenge: correctChallenge,
				Nonce:     correctNonce,
			},
			storageRespAllowed: true,
			storageRespErr:     nil,
			expectedCode:       models.SRCOK,
			expectedErr:        nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			strg.GetDelChallengeResponseAllowed = tc.storageRespAllowed
			strg.GetDelChallengeResponseErr = tc.storageRespErr

			actual, err := svr.checkChallenge(nil, tc.input)
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
				assert.Equal(t, tc.expectedCode, actual)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectedCode, actual)

		})
	}
}
