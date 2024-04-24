package auth

// Auth Response TLS Packet:
// Format    : [Version] [Payload]
// Payload   : [Response Code] [Protobuf Data]
//
// - Size (in bits):
//   - Response Code: 8
//   - SIDX: 32
//   - AES Key: 256

type response = uint8

const (
	responseUnknown       response = 0
	responseInternalError response = 1
	responseLoginTimeout  response = 2
	responseInvalidToken  response = 3
	responseInvalidServer response = 4
	responseLoginConflict response = 5
	responseLoginSuccess  response = 6
)
