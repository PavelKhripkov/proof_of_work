package protocol

type SeverMethod uint16

const (
	SMNoOp SeverMethod = iota
	SMGetQuote
)

type ServerResponseCode uint16

const (
	SRCUnknown ServerResponseCode = iota
	SRCOK
	SRCWrongNonce
	SRCUnknownProtocol
	SRCUnknownServerMethod
	SRCCantIdentifyClient
)
