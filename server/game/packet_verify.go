package game

import (
	"github.com/pemmel/gameserver/server"
)

// verify verifies the integrity of a packet using the session information retrieved from the
// provided session container. It selects the appropriate verification method based on the packet
// version and delegates the verification process to the corresponding version-specific function.
//
// Parameters:
//   - c: The session container used to retrieve session information.
//   - gpb: A general-purpose buffer used by the verification function to avoid memory allocation.
//     This buffer is reused across different verification methods to minimize memory usage.
//
// Returns:
//   - *handleT: A pointer to the handleT struct containing session information and decrypted payload,
//     or nil if verification fails.
func (p *packet) verify(c *server.SessionContainer, gpb []byte) *handleT {
	session := c.Get(p.sidx())
	if session == nil {
		return nil
	}

	v := p.version()
	if v != session.Version {
		return nil
	}

	switch v {
	case 1:
		return p.verifyV1(session, gpb)
	default:
		return nil
	}
}

// verifyV1 verifies the integrity of a packet using the provided session and payload data.
// It decrypts the payload using the session's cipher, checks for the correct sequence number,
// and constructs a handleT struct containing the session, address, request code, and payload.
// If any error occurs during decryption or validation, it returns nil.
//
// Parameters:
//   - s: The session used for decryption and validation.
//   - gpb: A temporary general-purpose buffer used to store the nonce value required for decryption.
//     If the length of gpb is less than the required nonce size, it will be resized
//     to accommodate the nonce value.
//
// Returns:
//   - *handleT: A pointer to the handleT struct containing session information and decrypted payload,
//     or nil if decryption or validation fails.
func (p *packet) verifyV1(s *server.Session, gpb []byte) *handleT {
	nonceSize := s.Cipher.NonceSize()
	nonce := parseNonce(gpb, nonceSize, p.sequence())

	cipher := p.payload()
	plain, err := s.Cipher.Open(cipher[:0], nonce, cipher, p.header())
	if err != nil || len(plain) < 1 {
		return nil
	}

	return &handleT{
		session:     s,
		addr:        p.addr,
		requestCode: plain[0],
		payload:     plain[1:],
	}
}

// fill fills in the buffer with the given value
func parseNonce(b []byte, nonceSize int, sequence uint32) []byte {
	if nonceSize < 4 {
		panic("nonceSize should not be less than 4")
	}

	if len(b) < nonceSize {
		b = make([]byte, nonceSize)
	}

	b[0] = byte(sequence >> 24)
	b[1] = byte(sequence >> 16)
	b[2] = byte(sequence >> 8)
	b[3] = byte(sequence)

	for i := 4; i < nonceSize; i++ {
		b[i] = 0
	}

	return b[:nonceSize]
}
