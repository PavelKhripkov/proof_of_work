package pow

import (
	"context"
	"encoding/binary"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"time"
)

// readClientRequest reads a client request and verifies that it satisfies basic protocol requirements.
func (s *pow) readClientRequest(conn io.Reader) ([]byte, error) {
	requestBytes := make([]byte, clientRequestSize+1)

	n, err := conn.Read(requestBytes)
	if n != clientRequestSize {
		if err != nil {
			return nil, err
		}
		return nil, ErrUnknownProtocol
	}

	if err != nil && err != io.EOF {
		return nil, err
	}

	requestBytes = requestBytes[:clientRequestSize]
	return requestBytes, nil
}

// validateHeaderHash checks that header meets protocol requirements and service settings
// and returns its hash if it is.
func (s *pow) validateHeaderHash(headerBytes []byte) ([]byte, error) {
	if len(headerBytes) != clientRequestHeaderSize {
		return nil, ErrUnknownProtocol
	}

	hash := s.hashcash.Hash(headerBytes)

	s.l.WithField("hash", hash).Debug("Header hash.")

	if !s.hashcash.ValidateHash(hash, uint(s.target)) {
		return nil, errors.Wrap(ErrInvalidHeader, "hash validation error")
	}

	return hash, nil
}

// prepareHeader returns client request header that satisfies protocol requirements and service settings.
func (s *pow) prepareHeader(ctx context.Context, localIP string) (*clientRequestHeader, error) {
	resource := net.ParseIP(localIP)
	if resource == nil {
		return nil, errors.Wrapf(ErrWrongValue, "couldn't parse localIP '%s'", localIP)
	}

	header := newClientRequestHeader(s.version, s.target, time.Now(), resource, 0)

	headerBytes := header.Marshal()
	nonce, err := s.hashcash.FindNonce(ctx, headerBytes[:clientRequestHeaderSize-8], uint(s.target))
	if err != nil {
		return nil, err
	}
	header.Counter = binary.LittleEndian.Uint64(nonce)

	if s.l.Logger.IsLevelEnabled(logrus.DebugLevel) {
		headerBytes = header.Marshal()
		hash := s.hashcash.Hash(headerBytes)

		s.l.WithField("hash", hash).Debug("Header hash.")
	}

	return header, nil
}
