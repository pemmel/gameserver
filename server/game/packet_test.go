package game

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"testing"

	"github.com/pemmel/gameserver/server"
)

func BenchmarkPacketVerifyV1(b *testing.B) {
	tests := []int{10, 50, 200, 500, 1400}

	for _, len := range tests {
		uid := uint(10)
		new := server.NewSessionV1
		s := server.SharedSession().NewSession(new, uid)
		p := newPacketAEAD(1, s, len)

		b.Run(fmt.Sprintf("Alloc-Len-%d", len), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = p.verify(server.SharedSession(), nil)
			}
		})

		b.Run(fmt.Sprintf("Len-%d", len), func(b *testing.B) {
			var gpb [20]byte
			for i := 0; i < b.N; i++ {
				_ = p.verify(server.SharedSession(), gpb[:])
			}
		})
	}
}

func newPacketAEAD(version uint8, session *server.Session, len int) packet {
	if session == nil {
		panic(session)
	}

	sidx := binary.BigEndian.AppendUint32(nil, session.Sidx)
	seqn := binary.BigEndian.AppendUint32(nil, rand.Uint32())
	code := byte(rand.Int())
	grpc := make([]byte, len)

	buffer := make([]byte, 0, minPacketLenV1+len)
	buffer = append(buffer, version)
	buffer = append(buffer, sidx...)
	buffer = append(buffer, seqn...)
	buffer = append(buffer, code)
	buffer = append(buffer, grpc...)

	nonce := make([]byte, session.Cipher.NonceSize())
	copy(nonce[:], seqn)

	session.Cipher.Seal(buffer[9:9], nonce, buffer[9:], buffer[0:9])

	return packet{
		addr: nil,
		data: buffer[:cap(buffer)],
	}
}
