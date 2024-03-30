package server

import (
	"context"
	"io"
	"pow/pkg/protocol"
)

type proto interface {
	// HandleClientRequest used by server to handle client request.
	HandleClientRequest(ctx context.Context, conn io.Reader, remoteHost string, checker func(context.Context, string, []byte) (bool, error)) (protocol.SeverMethod, error)
	// SendServerResponse used by server to send response to client.
	SendServerResponse(conn io.Writer, code protocol.ServerResponseCode, payload []byte) error
}

type storage interface {
	// GetSetHashByClient returns boolean value showing whether provided hash is allowed for provided client.
	// Same hash can be used by a client only once in specified period of time.
	// To fulfill this requirement, service stores that hash and checks if the client already used it recently.
	GetSetHashByClient(ctx context.Context, client string, hash []byte) (bool, error)
}
