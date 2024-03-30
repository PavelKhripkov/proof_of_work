package client

import (
	"context"
	"io"
	"pow/pkg/protocol"
)

type proto interface {
	// SendClientRequest used by client to send a request to remote server.
	SendClientRequest(ctx context.Context, conn io.Writer, clientID string, method protocol.SeverMethod) error
	// ReceiveServerResponse used by client to receive service response.
	ReceiveServerResponse(conn io.Reader) (protocol.ServerResponseCode, []byte, error)
}
