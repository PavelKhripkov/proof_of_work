package pow

import (
	"encoding/binary"
	"net"
	"pow/pkg/protocol"
	"time"
)

const (
	ClientRequestHeaderSize = 34
	ClientRequestSize       = 36

	ServerResponseMinLen = 2
)

// ClientRequestHeader represents client request header.
type ClientRequestHeader struct {
	Ver      byte
	Bits     byte
	Date     time.Time
	Resource net.IP
	Counter  uint64
}

// Marshal returns marshalled header.
func (s ClientRequestHeader) Marshal() []byte {
	l := 1 + 1 + 8 + 16 + 8
	res := make([]byte, l)

	res[0] = s.Ver
	res[1] = s.Bits
	binary.BigEndian.PutUint64(res[2:10], uint64(s.Date.UnixNano()))
	copy(res[10:26], s.Resource)
	binary.LittleEndian.PutUint64(res[26:34], s.Counter)

	return res
}

// Unmarshal parses a byte message into structure.
func (s *ClientRequestHeader) Unmarshal(msg []byte) error {
	if len(msg) != 34 {
		return ErrUnknownProtocol
	}

	s.Resource = make(net.IP, 16)

	s.Ver = msg[0]
	s.Bits = msg[1]
	s.Date = time.Unix(0, int64(binary.BigEndian.Uint64(msg[2:10])))
	copy(s.Resource, msg[10:26])
	s.Counter = binary.LittleEndian.Uint64(msg[26:34])

	return nil
}

// ClientRequest represents client request.
type ClientRequest struct {
	Header ClientRequestHeader
	Method protocol.SeverMethod
}

// Unmarshal parses a byte message into structure.
func (s *ClientRequest) Unmarshal(msg []byte) error {
	if len(msg) != 36 {
		return ErrUnknownProtocol
	}

	if err := s.Header.Unmarshal(msg[:34]); err != nil {
		return err
	}
	s.Method = protocol.SeverMethod(binary.BigEndian.Uint16(msg[34:36]))
	return nil
}

// Marshal returns marshalled client request.
func (s *ClientRequest) Marshal() []byte {
	res := make([]byte, ClientRequestSize)
	copy(res, s.Header.Marshal())
	binary.BigEndian.PutUint16(res[ClientRequestHeaderSize:], uint16(s.Method))

	return res
}

// ServerResponse represents server response.
type ServerResponse struct {
	Code protocol.ServerResponseCode
	Body []byte
}

// Marshal returns marshalled server response.
func (s ServerResponse) Marshal() []byte {
	res := make([]byte, 2+len(s.Body))

	binary.BigEndian.PutUint16(res, uint16(s.Code))
	copy(res[2:], s.Body)

	return res
}

// Unmarshal parses a byte message into structure.
func (s *ServerResponse) Unmarshal(msg []byte) error {
	if len(msg) < ServerResponseMinLen {
		return ErrUnknownProtocol
	}

	s.Body = make([]byte, len(msg)-2)

	s.Code = protocol.ServerResponseCode(binary.BigEndian.Uint16(msg[:2]))
	copy(s.Body, msg[2:])

	return nil
}
