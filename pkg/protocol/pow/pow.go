package pow

import (
	"bytes"
	"context"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"pow/pkg/protocol"
	"time"
)

// pow is a service that helps protect a server from DoS attacks.
// The service checks whether the client has completed some work before they can access the server's resources.
// The service uses a modified Hashcash algorithm under the hood.
type pow struct {
	l                  *logrus.Entry
	version            byte
	target             byte
	hashcash           hashcash
	headerTimeInterval time.Duration
}

// NewPoW create a new Proof of Work service.
func NewPoW(l *logrus.Entry, version byte, target byte, hashcash hashcash, headerTimeInterval time.Duration) *pow {
	return &pow{
		l:                  l,
		version:            version,
		target:             target,
		hashcash:           hashcash,
		headerTimeInterval: headerTimeInterval,
	}
}

// SendClientRequest used by client to send a request to remote server.
func (s *pow) SendClientRequest(ctx context.Context, conn io.Writer, localIP string, method protocol.SeverMethod) error {
	header, err := s.prepareHeader(ctx, localIP)
	if err != nil {
		return errors.Wrap(err, "couldn't prepare header")
	}

	req := clientRequest{
		Header: *header,
		Method: method,
	}

	s.l.WithField("request", req).Debug("Prepared client request.")

	reqBytes := req.Marshal()
	if err != nil {
		return errors.Wrap(err, "couldn't marshal client request")
	}

	_, err = conn.Write(reqBytes)
	if err != nil {
		return errors.Wrap(err, "couldn't write to connection")
	}

	return nil
}

// ReceiveServerResponse used by client to receive service response.
func (s *pow) ReceiveServerResponse(ctx context.Context, conn io.Reader) (protocol.ServerResponseCode, []byte, error) {
	var buf bytes.Buffer
	n, err := io.Copy(&buf, conn)
	if err != nil {
		return protocol.SRCUnknown, nil, errors.Wrap(err, "couldn't read from connection")
	}

	if n == 0 {
		return protocol.SRCUnknown, nil, ErrEmptyResponse
	}

	resp := serverResponse{}
	if err = resp.Unmarshal(buf.Bytes()); err != nil {
		return protocol.SRCUnknown, nil, errors.Wrap(err, "couldn't unmarshal server response")
	}

	return resp.Code, resp.Body, nil
}

// HandleClientRequest used by server to handle client request.
func (s *pow) HandleClientRequest(
	ctx context.Context,
	conn io.Reader,
	remoteHost string,
	checker func(context.Context, string, []byte) (bool, error),
) (protocol.SeverMethod, error) {

	requestBytes, err := s.readClientRequest(ctx, conn)
	if err != nil {
		return protocol.SMNoOp, errors.Wrap(err, "couldn't read client request")
	}

	hash, err := s.validateHeaderHash(ctx, requestBytes[:clientRequestHeaderSize])
	if err != nil {
		return protocol.SMNoOp, err
	}

	var request clientRequest
	if err = request.Unmarshal(requestBytes); err != nil {
		return protocol.SMNoOp, errors.Wrap(err, "couldn't unmarshal client request")
	}

	s.l.WithField("request", request).Debug("Got client request.")

	if err = s.validateHeader(request.Header, remoteHost); err != nil {
		return protocol.SMNoOp, err
	}

	allowed, err := checker(ctx, remoteHost, hash)
	if err != nil {
		return protocol.SMNoOp, err
	}

	if !allowed {
		return protocol.SMNoOp, ErrHashAlreadyUsed
	}

	return protocol.SMGetQuote, nil
}

// validateHeader checks if header settings meet server requirements.
func (s *pow) validateHeader(header clientRequestHeader, remoteHost string) error {
	if clientIP := header.Resource.String(); clientIP != remoteHost {
		return errors.Wrapf(ErrWrongClientID, "client specified their ID as %s, server identified client as %s", header.Resource.String(), remoteHost)
	}

	if s.version != header.Ver {
		return errors.Wrapf(ErrWrongVersion, "client version %d doesn't match server version %d", header.Ver, s.version)
	}

	if s.target < header.Bits {
		return errors.Wrapf(ErrTargetBits, "client's target bits %d less then server's target bits %d", header.Bits, s.target)
	}

	now := time.Now()
	if now.Add(-s.headerTimeInterval).After(header.Date) || now.Add(s.headerTimeInterval).Before(header.Date) {
		return errors.Wrapf(ErrInvalidHeaderTime, "header's time is not within interval +- %v from current time", s.headerTimeInterval)
	}

	return nil
}

// SendServerResponse used by server to send response to client.
func (s *pow) SendServerResponse(ctx context.Context, conn io.Writer, code protocol.ServerResponseCode, payload []byte) error {
	resp := serverResponse{
		Code: code,
		Body: payload,
	}

	respBytes := resp.Marshal()

	_, err := conn.Write(respBytes)
	if err != nil {
		return errors.Wrap(err, "couldn't write to connection")
	}

	return nil
}
