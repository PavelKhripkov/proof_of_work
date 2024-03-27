package client

import (
	"context"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net"
	"pow/internal/protocol"
)

const tcp = "tcp"

type client struct {
	l     *logrus.Entry
	proto proto
	done  chan struct{}
}

func NewClient(l *logrus.Entry, proto proto) *client {
	return &client{
		l:     l,
		proto: proto,
		done:  make(chan struct{}),
	}
}

type DoParams struct {
	RemoteAddr string
	Method     protocol.SeverMethod
}

func (s *client) Do(ctx context.Context, params DoParams) ([]byte, error) {
	conn, err := net.Dial(tcp, params.RemoteAddr)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't connect to host")
	}
	defer func() {
		if err := conn.Close(); err != nil {
			s.l.Error(err)
		}
	}()

	hostPort := conn.LocalAddr().String()
	localIP, _, err := net.SplitHostPort(hostPort)
	if err != nil {
		return nil, err
	}

	if err = s.proto.SendClientRequest(ctx, conn, localIP, params.Method); err != nil {
		return nil, err
	}

	code, payload, err := s.proto.ReceiveServerResponse(ctx, conn)
	if err != nil {
		return nil, err
	}

	if code != protocol.SRCOK {
		return nil, errors.Wrapf(ErrServerErrorCode, "code %d", code)
	}

	return payload, nil
}

func (s *client) Done() {
	<-s.done
}
