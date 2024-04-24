package game

import (
	"encoding/binary"
	"net"
)

// Game UDP Packet Format:
// Hierarchy : [Header] [Payload]
// Header    : [Version] [SIDX] [Sequence Number]
// Payload   : [Request Code] [Protobuf Data]
// Version 1 : [1] [SIDX] [Sequence Number] [GCM Auth Tag] [Payload]
//
// - Encryption:
//     Request Code and gRPC Payload are encrypted together using AES256-GCM with the
// 		 Sequence Number as the nonce.
//
// - Size (in bits):
//   - Version: 8
//   - SIDX: 32
//   - Sequence Number: 32
//   - Request Code: 8
//   - GCM Auth Tag: 128
//   - Protobuf Data: Variable
//
// The client encrypts the determined field with the provided ECDH public key.
// The server validates the request with its server-side ECDH private key corresponding
// to the packet SIDX requested by the client. Then, the server decodes the gRPC payload
// according to the request code.

const (
	gcmTagLen      int = 16
	versionLen     int = 8 / 8
	sidxLen        int = 32 / 8
	sequenceNbLen  int = 32 / 8
	requestCodeLen int = 8 / 8
	minPacketLenV1 int = versionLen + sidxLen + sequenceNbLen + gcmTagLen + requestCodeLen

	versionBeginPos    int = 0
	versionEndPos      int = versionBeginPos + versionLen
	sidxBeginPos       int = versionEndPos
	sidxEndPos         int = sidxBeginPos + sidxLen
	sequenceNbBeginPos int = sidxEndPos
	sequenceNbEndPos   int = sequenceNbBeginPos + sequenceNbLen
	payloadBeginPos    int = sequenceNbEndPos
)

func packetMeaningful(b []byte) bool {
	if len(b) == 0 {
		return false
	}
	v := b[0]
	switch v {
	case 1:
		return len(b) > minPacketLenV1
	default:
		return false
	}
}

type packet struct {
	addr *net.UDPAddr
	data []byte
}

func (p *packet) header() []byte {
	return p.data[0:sequenceNbEndPos]
}

func (p *packet) payload() []byte {
	return p.data[payloadBeginPos:]
}

func (p *packet) version() uint8 {
	return p.data[versionBeginPos]
}

func (p *packet) sidx() uint32 {
	s := p.data[sidxBeginPos:sidxEndPos]
	return binary.BigEndian.Uint32(s)
}

func (p *packet) sequence() uint32 {
	s := p.data[sequenceNbBeginPos:sequenceNbEndPos]
	return binary.BigEndian.Uint32(s)
}
