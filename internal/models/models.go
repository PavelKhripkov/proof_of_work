package models

// SeverMethod is the type that is used by client to request some server resources.
type SeverMethod uint16

const (
	// SMNoOp is No Operation.
	SMNoOp SeverMethod = iota

	// SMGetChallenge method to get a new challenge from server.
	SMGetChallenge

	// SMGetQuote method to get a quote.
	SMGetQuote
)

// ServerResponseCode is the type that is used by server to inform client about status of client's request.
type ServerResponseCode uint16

const (
	// SRCInternalError represents errors not related to clients request.
	SRCInternalError ServerResponseCode = iota

	// SRCOK successful response result.
	SRCOK

	// SRCWrongChallenge returned when client specified not known or spoiled challenge.
	SRCWrongChallenge

	// SRCWrongNonce returned when provided nonce doesn't lead to correct target.
	SRCWrongNonce

	// SRCUnknownMethod returned when specified method is unknown.
	SRCUnknownMethod
)

// ClientRequest represents client request.
type ClientRequest struct {
	Method    SeverMethod
	Challenge []byte
	Nonce     []byte
}

// ServerResponse represents server response.
type ServerResponse struct {
	Code      ServerResponseCode
	Body      []byte
	Challenge []byte
	Target    byte
}
