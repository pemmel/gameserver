package test

import (
	"fmt"
	"testing"

	"github.com/pemmel/gameserver/server"
)

func BenchmarkGenerateClientKey(b *testing.B) {
	tests := []int{8, 16, 32, 64, 128}

	for _, t := range tests {
		b.Run(fmt.Sprintf("Len-%d", t), func(b *testing.B) {
			buffer := make([]byte, t)
			for i := 0; i < b.N; i++ {
				server.GenerateKey(buffer)
			}
		})
	}
}

func BenchmarkNewSession(b *testing.B) {
	b.Run("V1", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = server.SharedSession().NewSession(server.NewSessionV1, uint(i))
		}
	})
}
