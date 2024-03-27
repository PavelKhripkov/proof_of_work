package client

import (
	"context"
	"io"
	"pow/internal/protocol"
)

type proto interface {
	SendClientRequest(ctx context.Context, conn io.Writer, clientID string, method protocol.SeverMethod) error
	ReceiveServerResponse(ctx context.Context, conn io.Reader) (protocol.ServerResponseCode, []byte, error)
}
