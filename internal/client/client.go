package client

import (
	"context"
	"encoding/gob"
	"fmt"
	"net"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"pow/internal/models"
)

const tcp = "tcp"

// Client represents a service that is making some work before asking a server for some resources.
type Client struct {
	l        *logrus.Entry
	hashcash hashcash
}

// NewClient returns new client instance.
func NewClient(l *logrus.Entry, h hashcash) *Client {
	return &Client{
		l:        l,
		hashcash: h,
	}
}

// GetQuotesParams contains params for GetQuotes func.
type GetQuotesParams struct {
	ServerAddr            string
	QuotesDesired         byte
	ProvideWrongChallenge bool
	ProvideWrongNonce     bool
}

// GetQuotes returns specified amount of quotes received from the server.
func (s *Client) GetQuotes(ctx context.Context, params GetQuotesParams) ([]string, error) {
	conn, err := net.Dial(tcp, params.ServerAddr)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't connect to host")
	}
	defer func() {
		if err := conn.Close(); err != nil {
			s.l.Error(err)
		}
	}()

	encoder := gob.NewEncoder(conn)
	decoder := gob.NewDecoder(conn)
	var res []string
	var nonce []byte

	requestGetChallenge := models.ClientRequest{
		Method: models.SMGetChallenge,
	}
	response := models.ServerResponse{}

	for i := byte(0); i < params.QuotesDesired; i++ {

		if err = encoder.Encode(requestGetChallenge); err != nil {
			return res, err
		}

		if err = decoder.Decode(&response); err != nil {
			return res, err
		}

		if params.ProvideWrongChallenge {
			response.Challenge = append(response.Challenge, 24)
		}

		nonce, err = s.hashcash.FindNonce(ctx, response.Challenge, uint(response.Target))
		if err != nil {
			return res, err
		}

		if params.ProvideWrongNonce {
			nonce = append(nonce, 24)
		}

		requestGetQuote := models.ClientRequest{
			Method:    models.SMGetQuote,
			Challenge: response.Challenge,
			Nonce:     nonce,
		}

		if err = encoder.Encode(requestGetQuote); err != nil {
			return res, err
		}

		if err = decoder.Decode(&response); err != nil {
			return res, err
		}

		if response.Code != models.SRCOK {
			res = append(res, fmt.Sprintf("Server responded with error code: %d", response.Code))
			continue
		}

		res = append(res, string(response.Body))
	}

	return res, nil
}
