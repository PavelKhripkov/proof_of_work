package server

import (
	"context"
	"crypto/rand"
	"encoding/gob"
	"io"
	"net"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"pow/internal/models"
)

const tcp = "tcp"

// Server is a web server that uses some PoW services to protect itself from DoS attacks.
type Server struct {
	l             *logrus.Entry
	storage       storage
	hashcash      hashcash
	respTimeout   time.Duration
	challengeSize byte
	targetBits    byte
	challengeTTL  time.Duration
	done          chan struct{}
}

// NewServerParams contains parameters for NewServer function.
type NewServerParams struct {
	Logger        *logrus.Entry
	Storage       storage
	Hashcash      hashcash
	RespTimeout   time.Duration
	ChallengeSize byte
	TargetBits    byte
	ChallengeTTL  time.Duration
}

// NewServer returns a new server instance.
func NewServer(p NewServerParams) *Server {
	return &Server{
		l:             p.Logger,
		storage:       p.Storage,
		hashcash:      p.Hashcash,
		respTimeout:   p.RespTimeout,
		challengeSize: p.ChallengeSize,
		targetBits:    p.TargetBits,
		challengeTTL:  p.ChallengeTTL,
		done:          make(chan struct{}),
	}
}

// Run starts the server.
func (s *Server) Run(ctx context.Context, addr string) error {
	var lstn net.Listener

	go func() {
		<-ctx.Done()
		if lstn != nil {
			if err := lstn.Close(); err != nil {
				err = errors.Wrap(err, "error occurred on listener close")
				s.l.Error(err)
			}
		}

		close(s.done)
	}()

	lstn, err := net.Listen(tcp, addr)
	if err != nil {
		return errors.Wrap(err, "couldn't init listener")
	}

	for {
		conn, err := lstn.Accept()
		if err != nil {
			return err
		}

		go func() {
			ctxServe, cancel := context.WithTimeout(ctx, s.respTimeout)
			defer cancel()
			s.serveConn(ctxServe, conn)
		}()
	}
}

// serveConn serves client's connection.
func (s *Server) serveConn(ctx context.Context, conn net.Conn) {
	var err error

	defer func() {
		if err != nil {
			s.l.Error(err)
		}
		if err = conn.Close(); err != nil {
			err = errors.Wrap(err, "couldn't close connection")
			s.l.Error(err)
		}
	}()

	deadline, _ := ctx.Deadline()
	if err = conn.SetDeadline(deadline); err != nil {
		err = errors.Wrap(err, "couldn't set connection deadline")
		return
	}

	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)
	req := models.ClientRequest{}
	resp := models.ServerResponse{}

	for {
		err = decoder.Decode(&req)
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return
		}

		resp, err = s.serveRequest(ctx, req)
		if err != nil {
			return
		}

		err = encoder.Encode(resp)
		if err != nil {
			return
		}

		// TODO Deadline could be refreshed to extend keep alive time.
	}
}

// serveRequest routes client requests to an appropriate handler.
func (s *Server) serveRequest(ctx context.Context, request models.ClientRequest) (models.ServerResponse, error) {
	switch request.Method {
	case models.SMGetChallenge:
		return s.getChallengeHandler(ctx)
	case models.SMGetQuote:
		return s.getQuoteHandler(ctx, request)
	default:
		return models.ServerResponse{Code: models.SRCUnknownMethod}, nil
	}
}

// getChallengeHandler returns a new challenge.
func (s *Server) getChallengeHandler(ctx context.Context) (resp models.ServerResponse, err error) {
	challenge := make([]byte, s.challengeSize)
	_, err = rand.Read(challenge)
	if err != nil {
		resp.Code = models.SRCInternalError
		return
	}

	if err = s.storage.StoreChallenge(ctx, challenge, s.challengeTTL); err != nil {
		resp.Code = models.SRCInternalError
		return
	}

	resp.Challenge = challenge
	resp.Target = s.targetBits
	resp.Code = models.SRCOK

	return
}

// getQuoteHandler returns a quote.
func (s *Server) getQuoteHandler(ctx context.Context, request models.ClientRequest) (models.ServerResponse, error) {
	code, err := s.checkChallenge(ctx, request)
	if err != nil {
		return models.ServerResponse{Code: models.SRCInternalError}, err
	}

	if code != models.SRCOK {
		return models.ServerResponse{Code: code}, nil
	}

	return models.ServerResponse{
		Code:      models.SRCOK,
		Body:      []byte(randQuote()),
		Challenge: nil,
		Target:    0,
	}, nil
}

// checkChallenge checks whether the challenge valid.
func (s *Server) checkChallenge(ctx context.Context, request models.ClientRequest) (models.ServerResponseCode, error) {
	exists, err := s.storage.GetDelChallenge(ctx, request.Challenge)
	if err != nil {
		return models.SRCInternalError, err
	}

	if !exists {
		return models.SRCWrongChallenge, nil
	}

	hash := s.hashcash.Hash(append(request.Challenge, request.Nonce...))

	valid := s.hashcash.ValidateHash(hash, uint(s.targetBits))
	if !valid {
		return models.SRCWrongNonce, nil
	}

	return models.SRCOK, nil
}

// Done waits the server is stopped gracefully.
func (s *Server) Done() {
	<-s.done
}
