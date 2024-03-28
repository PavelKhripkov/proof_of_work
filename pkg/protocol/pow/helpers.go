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
func (s *pow) readClientRequest(ctx context.Context, conn io.Reader) ([]byte, error) {
	requestBytes := make([]byte, ClientRequestSize+1)

	n, err := conn.Read(requestBytes)

	if n != ClientRequestSize {
		return nil, ErrUnknownProtocol
	}

	if err != nil && err != io.EOF {
		return nil, err
	}

	requestBytes = requestBytes[:ClientRequestSize]

	return requestBytes, nil
}

// validateHeaderHash checks that header meets protocol requirements and service settings.
func (s *pow) validateHeaderHash(ctx context.Context, headerBytes []byte) ([]byte, error) {
	if len(headerBytes) != ClientRequestHeaderSize {
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
func (s *pow) prepareHeader(ctx context.Context, localIP string) (*ClientRequestHeader, error) {
	resource := net.ParseIP(localIP)
	if resource == nil {
		return nil, errors.Wrapf(ErrWrongValue, "couldn't parse localIP '%s'", localIP)
	}

	header := ClientRequestHeader{
		Ver:      s.version,
		Bits:     s.target,
		Date:     time.Now(),
		Resource: resource,
		Counter:  0,
	}

	headerBytes := header.Marshal()
	nonce, err := s.hashcash.FindNonce(ctx, headerBytes[:ClientRequestHeaderSize-8], uint(s.target))
	if err != nil {
		return nil, err
	}
	header.Counter = binary.LittleEndian.Uint64(nonce)

	if s.l.Logger.IsLevelEnabled(logrus.DebugLevel) {
		headerBytes = header.Marshal()
		hash := s.hashcash.Hash(headerBytes)

		s.l.WithField("hash", hash).Debug("Header hash.")
	}

	return &header, nil
}
