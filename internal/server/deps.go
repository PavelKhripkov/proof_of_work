package server

import (
	"context"
	"io"
	"pow/internal/protocol"
)

type proto interface {
	HandleClientRequest(ctx context.Context, conn io.Reader, remoteHost string, checker func(context.Context, string, []byte) (bool, error)) (protocol.SeverMethod, error)
	SendServerResponse(ctx context.Context, conn io.Writer, code protocol.ServerResponseCode, payload []byte) error
}

type storage interface {
	GetSetHashByClient(ctx context.Context, client string, hash []byte) (bool, error)
}
