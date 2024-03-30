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
	// SRCUnknownError represents unknown response or response absence.
	SRCUnknownError ServerResponseCode = iota
	// SRCOK successful response result.
	SRCOK
	// SRCUnknownProtocol returned if server cannot recognize protocol.
	SRCUnknownProtocol
	// SRCCannotIdentifyClient returned when server cannot identity client.
	SRCCannotIdentifyClient
	// SRCWrongVersion returned when specified wrong version of protocol.
	SRCWrongVersion
	// SRCWrongTargetBits returned when specified wrong amount of target bits.
	SRCWrongTargetBits
	// SRCHashAlreadyUsed returned when hash already used by client.
	SRCHashAlreadyUsed
	// SRCInvalidHeader returned when provided wrong header.
	SRCInvalidHeader
	// SRCInvalidHeaderTime returned when header time is out of range.
	SRCInvalidHeaderTime
)
