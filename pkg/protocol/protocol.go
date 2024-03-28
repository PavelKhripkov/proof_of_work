package protocol

// SeverMethod is the type that is used by client to request some server resources.
type SeverMethod uint16

const (
	// SMNoOp is No Operation.
	SMNoOp SeverMethod = iota
	// SMGetQuote method to get a quote.
	SMGetQuote
)

// ServerResponseCode is the type that is used by server to inform client about status of client's request.
type ServerResponseCode uint16

const (
	// SRCUnknown represents unknown response or response absence.
	SRCUnknown ServerResponseCode = iota
	// SRCOK successful response result.
	SRCOK
	// SRCWrongNonce returned if client provided wrong nonce.
	SRCWrongNonce
	// SRCUnknownProtocol returned if server cannot recognize protocol.
	SRCUnknownProtocol
	// SRCCantIdentifyClient returned when server cannot identity client.
	SRCCantIdentifyClient
)
