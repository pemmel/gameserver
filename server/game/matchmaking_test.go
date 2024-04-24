package game

// given a := []int{1, 2, 2, 3, 4, 1}
// all of the combination must be unique.
// there's no subset in first pair is appearing in second pair
// must prioritize the smallest index: a[0], a[1], a[...]
// combinations k=any target=5, listed element is type of index:
// - [0 1 2]
// - [0 3 5]
// - [1 2 5]
// - [1 3]
// - [2 3]
// - [0 4]
// - [4 5]
// expected:
// [0 1 2] [4 5]
// invalid (subset index 3 reappearing):
// [1 3] [2 3]
// invalid (not prioritizing smallest index):
// [2 3] [4 5]

import (
	"fmt"
	"testing"
	"time"

	"github.com/bytedance/gopkg/lang/fastrand"
)

func TestMatchmaking(t *testing.T) {
	a := []int{1, 2, 2, 3, 4, 4}
	for i, v := range a {
		r := LobbyRoom{
			Idx:      uint32(i),
			HostSidx: uint32(v),
			Guests:   make([]LobbyGuest, v-1),
		}
		mmQueue.Insert(r)
	}
	findmatch()
	for i := mmQueue.next; i != nil; i = i.next {
		v := i.Value
		fmt.Printf("Idx:%d, Cnt:%d\n", v.Idx, v.PlayerCount())
	}
}

func TestMatchmakingIndefinitely(t *testing.T) {
	go matchmaking()
	for idx := uint32(0); true; idx++ {
		cnt := fastrand.Uint32() % 3
		mmQueue.Insert(LobbyRoom{
			Idx:    idx,
			Guests: make([]LobbyGuest, cnt),
		})
		time.Sleep(time.Millisecond)
	}
}

func BenchmarkBorrow(b *testing.B) {
	tests := []int{5, 10, 15, 20, 30, 45, 60, 90}
	for _, n := range tests {
		b.Run(fmt.Sprintf("Len-%d", n), func(b *testing.B) {
			buf := make([]*LlistNode[LobbyRoom], 0, n)
			for i := 0; i < n; i++ {
				mmQueue.Insert(LobbyRoom{})
			}
			for i := 0; i < b.N; i++ {
				mmQueue.Borrow(n, &buf)
			}
		})
	}
}

func BenchmarkUniqueElement(b *testing.B) {
	b.Run("Worst-5", func(b *testing.B) {
		a1 := []int{0, 1, 2, 3, 4}
		a2 := []int{5, 6, 7, 8, 9}
		for i := 0; i < b.N; i++ {
			if !compareUnique(a1, a2) {
				b.FailNow()
			}
		}
	})
}

func BenchmarkMatchmaking(b *testing.B) {
	tests := []int{2, 5, 10, 15, 30, 50, 100, 200}
	for _, t := range tests {
		a := make([]*LlistNode[LobbyRoom], t)
		for i := 0; i < t; i++ {
			v := fastrand.Uint32() % 5
			a[i] = &LlistNode[LobbyRoom]{
				next:  nil,
				Value: LobbyRoom{Guests: make([]LobbyGuest, v)},
			}
		}

		b.Run(fmt.Sprintf("Queue-%d", t), func(b *testing.B) {
			var buf [10]int
			for i := 0; i < b.N; i++ {
				c := combinations(a)
				_ = mergeUnique(c, &buf)
			}
		})
	}
}
