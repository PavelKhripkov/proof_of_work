package pow

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net"
	"pow/pkg/protocol"
	"testing"
	"time"
)

func TestClientRequestHeaderMarshalUnmarshal(t *testing.T) {
	header := newClientRequestHeader(
		32,
		34,
		time.Now(),
		net.IP{192, 168, 1, 10},
		12947832)

	headerBytes := header.Marshal()
	require.Equal(t, clientRequestHeaderSize, len(headerBytes))

	actual := clientRequestHeader{}
	err := actual.Unmarshal(headerBytes)
	require.NoError(t, err)
	assert.Equal(t, *header, actual)
}

func TestClientRequestMarshalUnmarshal(t *testing.T) {
	header := newClientRequestHeader(
		0,
		20,
		time.Time{},
		nil,
		0)

	request := clientRequest{
		Header: *header,
		Method: protocol.SMNoOp,
	}

	requestBytes := request.Marshal()
	require.Equal(t, clientRequestSize, len(requestBytes))

	actual := clientRequest{}
	err := actual.Unmarshal(requestBytes)
	require.NoError(t, err)
	assert.Equal(t, request, actual)
}

func TestServerResponseMarshalUnmarshal(t *testing.T) {
	response := serverResponse{
		Code: protocol.SRCHashAlreadyUsed,
		Body: []byte("some payload of response body"),
	}

	responseBytes := response.Marshal()
	actual := serverResponse{}
	err := actual.Unmarshal(responseBytes)
	require.NoError(t, err)
	assert.Equal(t, response, actual)
}
