package server

import (
	"context"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net"
	"pow/pkg/protocol"
	"time"
)

// server is a web server that uses some PoW services to protect itself from DoS attacks.
type server struct {
	l           *logrus.Entry
	storage     storage
	proto       proto
	respTimeout time.Duration
	done        chan struct{}
}

// NewServer returns a new server instance.
func NewServer(l *logrus.Entry, storage storage, proto proto, respTimeout time.Duration) *server {
	return &server{
		l:           l,
		storage:     storage,
		proto:       proto,
		respTimeout: respTimeout,
		done:        make(chan struct{}),
	}
}

// Run starts the server.
func (s *server) Run(ctx context.Context, addr net.Addr) error {
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

	lstn, err := net.Listen(addr.Network(), addr.String())
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
func (s *server) serveConn(ctx context.Context, conn net.Conn) {
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

	remoteHostPort := conn.RemoteAddr().String()
	if remoteHostPort == "" {
		err = s.sendError(ctx, conn, protocol.SRCCantIdentifyClient)
		return
	}

	remoteHost, _, err := net.SplitHostPort(remoteHostPort)
	if err != nil {
		return
	}

	s.l.Debug("Serving connection from remote client ", remoteHost)

	method, err := s.proto.HandleClientRequest(ctx, conn, remoteHost, s.storage.GetSetHashByClient)
	if err != nil {
		return
	}

	switch method {
	case protocol.SMNoOp:
		return
	case protocol.SMGetQuote:
		err = s.sendQuote(ctx, conn)
	default:
		err = s.sendError(ctx, conn, protocol.SRCUnknown)
	}

	if err != nil {
		return
	}

	return
}

// sendQuote sends a quote into connection as a response to a client.
func (s *server) sendQuote(ctx context.Context, conn net.Conn) error {
	payload := randQuote()

	if err := s.proto.SendServerResponse(ctx, conn, protocol.SRCOK, []byte(payload)); err != nil {
		return errors.Wrap(err, "couldn't send server response")
	}

	return nil
}

// sendError sends an error response to client.
func (s *server) sendError(ctx context.Context, conn net.Conn, code protocol.ServerResponseCode) error {
	if err := s.proto.SendServerResponse(ctx, conn, protocol.SRCOK, nil); err != nil {
		return errors.Wrap(err, "couldn't send server response")
	}

	return nil
}

func (s *server) Done() {
	<-s.done
}
