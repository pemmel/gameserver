package game

import (
	"fmt"
	"time"
)

const (
	mmWaitInterval        = time.Second
	mmBorrowStride    int = 10
	mmTeamSize        int = 2
	mmPlayerPerTeam   int = 5
	mmTotalPlayerSize int = mmTeamSize * mmPlayerPerTeam
)

var (
	mmQueue   *LlistHead[LobbyRoom]
	mmCancels []uint32
)

func init() {
	mmQueue = NewLList[LobbyRoom]()
	mmCancels = make([]uint32, 10)
}

func matchmaking() {
	for {
		m, b := findmatch()
		fmt.Printf("Lobbies Matched: %d | Queue Borrowed: %d | Queue Available: %d\n", m, b, b-m)
		time.Sleep(mmWaitInterval)
	}
}

// findmatch performs matchmaking by searching for the required combination of 5v5 matches.
// It borrows lobby queues and attempts to group and merge them if the required combination is found.
// The function utilizes a buffer to store the borrowed lobby queues, and when the buffer is not large
// enough to accommodate them, it dynamically allocates additional space.
// Parameters:
//
//	mmStride: The number of lobby queues to process at a time during matchmaking.
//	mmQueue: The queue manager responsible for managing the lobby queues.
//
// Returns:
//
//	mn: The count of how many lobbies are merged to form matches.
//	bn: The amount of borrowed lobby queues.
func findmatch() (mn int, bn int) {
	const bcap = 3 * mmBorrowStride
	var mrg [mmTotalPlayerSize]int      // Stores merged lobby for matchmaking
	var buf [bcap]*LlistNode[LobbyRoom] // Buffer for storing borrowed lobby

	// Matchmaking process.
	// Loop until match cannot be found.
	r := buf[:0]
	for {
		// Ensure buffer capacity is sufficient to store borrowed lobby queues
		if cap(r) < len(r)+mmBorrowStride {
			t := make([]*LlistNode[LobbyRoom], len(r)+mmBorrowStride)
			copy(t, r)
			r = t[:len(r)]
		}
		// Borrow lobby queues from the queue manager
		// Return unused borrowed lobby queues if matchmaking is unsuccessful
		w := r[len(r):]
		b := mmQueue.Borrow(mmBorrowStride, &w)
		if b == 0 {
			if len(r) != 0 {
				head := r[0]
				tail := r[len(r)-1]
				mmQueue.Return(head, tail)
			}
			return
		}
		if len(r) != 0 {
			r[len(r)-1].next = w[0]
		}
		r = r[:len(r)+b]
		bn += b
		// we try matching all of the borrowed lobbies to form a match.
		// If no lobbies can be paired, try borrowing another lobby.
	matching:
		c := combinations(r)
		m := mergeUnique(c, &mrg)
		if m == 0 {
			continue
		}
		mn += m
		// Remove unused lobbies from the list by rechaining
		l := make([]LobbyRoom, m)
		for i, j := range mrg[:m] {
			l[i] = r[j].Value
			r[j] = nil
		}
		_ = NewMatchConfig(l)
		// Update the r buffer with latest structure
		x := r[:0]
		p := &LlistNode[LobbyRoom]{next: nil}
		for _, n := range r {
			if n != nil {
				x = append(x, n)
				p.next = n
				p = n
			}
		}
		p.next = nil // p is now the tail. tail.next should be nil
		r = x
		// There's chance that we can still find another match.
		// Thus we go calculate the possibility again without the
		// needs borrow from the queue manager.
		if len(r) >= 4 {
			goto matching
		}
	}
}

func combinations(a []*LlistNode[LobbyRoom]) [][]int {
	var result [][]int
	var curr [mmPlayerPerTeam]int
	backtrack(a, 0, 0, curr[:0], &result)
	return result
}

func backtrack(a []*LlistNode[LobbyRoom], start, sum int, current []int, result *[][]int) {
	if sum == mmPlayerPerTeam {
		*result = append(*result, append([]int{}, current...))
		return
	}

	if sum > mmPlayerPerTeam {
		return
	}

	for i := start; i < len(a); i++ {
		newSum := sum + a[i].Value.PlayerCount()
		if newSum <= mmPlayerPerTeam {
			current = append(current, i)
			backtrack(a, i+1, newSum, current, result)
			current = current[:len(current)-1]
		}
	}
}

func mergeUnique[T comparable](result [][]T, r *[mmTotalPlayerSize]T) int {
	for i := 0; i < len(result); i++ {
		a1 := result[i]
		for j := i + 1; j < len(result); j++ {
			a2 := result[j]
			if compareUnique(a2, a1) {
				w := (*r)[:0]
				w = append(w, a1...)
				w = append(w, a2...)
				return len(w)
			}
		}
	}
	return 0
}

func compareUnique[T comparable](a, b []T) bool {
	for _, e1 := range a {
		for _, e2 := range b {
			if e1 == e2 {
				return false
			}
		}
	}
	return true
}
