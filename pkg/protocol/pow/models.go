package pow

import (
	"encoding/binary"
	"net"
	"pow/pkg/protocol"
	"time"
)

const (
	clientRequestHeaderSize = 34
	clientRequestSize       = 36

	serverResponseMinLen = 2
)

// clientRequestHeader represents client request header.
type clientRequestHeader struct {
	Ver      byte
	Bits     byte
	Date     time.Time
	Resource net.IP
	Counter  uint64
}

// newClientRequestHeader returns new clientRequestHeader created according to fixed-size specs of current protocol.
func newClientRequestHeader(ver, bits byte, date time.Time, resource net.IP, counter uint64) *clientRequestHeader {
	res := make([]byte, 16)
	copy(res[16-len(resource):], resource)
	return &clientRequestHeader{
		Ver:      ver,
		Bits:     bits,
		Date:     time.Unix(0, date.UnixNano()),
		Resource: res,
		Counter:  counter,
	}
}

// Marshal returns marshalled header.
func (s clientRequestHeader) Marshal() []byte {
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
func (s *clientRequestHeader) Unmarshal(msg []byte) error {
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

// clientRequest represents client request.
type clientRequest struct {
	Header clientRequestHeader
	Method protocol.SeverMethod
}

// Unmarshal parses a byte message into structure.
func (s *clientRequest) Unmarshal(msg []byte) error {
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
func (s *clientRequest) Marshal() []byte {
	res := make([]byte, clientRequestSize)
	copy(res, s.Header.Marshal())
	binary.BigEndian.PutUint16(res[clientRequestHeaderSize:], uint16(s.Method))

	return res
}

// serverResponse represents server response.
type serverResponse struct {
	Code protocol.ServerResponseCode
	Body []byte
}

// Marshal returns marshalled server response.
func (s serverResponse) Marshal() []byte {
	res := make([]byte, 2+len(s.Body))

	binary.BigEndian.PutUint16(res, uint16(s.Code))
	copy(res[2:], s.Body)

	return res
}

// Unmarshal parses a byte message into structure.
func (s *serverResponse) Unmarshal(msg []byte) error {
	if len(msg) < serverResponseMinLen {
		return ErrUnknownProtocol
	}

	s.Body = make([]byte, len(msg)-2)

	s.Code = protocol.ServerResponseCode(binary.BigEndian.Uint16(msg[:2]))
	copy(s.Body, msg[2:])

	return nil
}
