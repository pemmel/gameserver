package game

import (
	"fmt"
	"time"
)

const (
	mmStride int = 20

	mmTeamSize        int = 2
	mmPlayerPerTeam   int = 5
	mmTotalPlayerSize int = mmTeamSize * mmPlayerPerTeam

	mmMaxWaitTime        = 5 * time.Second
	mmMinWaitTime        = 1 * time.Second
	mmLowQueueCount  int = 10
	mmHighQueueCount int = 50
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
		m := 0
		n := 0
		a := 0
		for mi := 1; mi != 0; {
			mi, a = findmatch()
			m += mi
			n += a
		}
		fmt.Printf("Lobbies Matched: %d | Queue Borrowed: %d | Queue Available: %d\n", m, n, a)
		time.Sleep(time.Second)
	}
}

func waitTimeDuration(cnt int) time.Duration {
	switch {
	case cnt <= mmLowQueueCount:
		return mmMaxWaitTime
	case cnt >= mmHighQueueCount:
		return mmMinWaitTime
	default:
		waitFraction := float64(cnt-mmLowQueueCount) / float64(mmHighQueueCount-mmLowQueueCount)
		waitDuration := float64(mmMaxWaitTime-mmMinWaitTime) * waitFraction
		return mmMinWaitTime + time.Duration(waitDuration)
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
//	m: The count of how many lobbies are merged to form matches.
//	n: The amount of borrowed lobby queues.
func findmatch() (m int, n int) {
	var mrg [mmTotalPlayerSize]int              // Stores merged lobby for matchmaking
	var buf [1 + mmStride]*LlistNode[LobbyRoom] // Buffer for storing borrowed lobby

	r := buf[0:1]
	r[0] = &LlistNode[LobbyRoom]{next: nil}

	// Matchmaking process.
	// Loop until a match is found or no more lobbies are available
	for m == 0 {
		// Ensure buffer capacity is sufficient to store borrowed lobby queues
		if cap(r) < 1+n+mmStride {
			t := make([]*LlistNode[LobbyRoom], 1+n+mmStride)
			copy(t, r)
			r = t
		}
		// Borrow lobby queues from the queue manager
		// Return unused borrowed lobby queues if matchmaking is unsuccessful
		w := r[1+n : 1+n]
		g := mmQueue.Borrow(mmStride, &w)
		if g == 0 {
			if n != 0 {
				head := r[0].next
				tail := r[n]
				mmQueue.Return(head, tail)
			}
			return
		}
		// Link borrowed lobby queues to the buffer
		r[n].next = w[0]
		n += g
		r = r[:1+n]
		// Determine potential matches and merge lobbies if possible
		c := combinations(r[1:])
		m = mergeUnique(c, &mrg)
	}

	// Remove unused lobbies from the list by rechaining
	for _, i := range mrg[:m] {
		node := r[i+1]
		r[i].next = node.next
		r[i+1] = r[i]
	}

	// Return unused borrowed lobby queues after rechaining
	head := r[0].next
	for tail := head; tail != nil; tail = tail.next {
		if tail.next == nil {
			mmQueue.Return(head, tail)
			break
		}
	}

	return
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
	var a1 [mmPlayerPerTeam]T
	for i := 0; i < len(result); i++ {
		n := copy(a1[:], result[i])
		for j := i + 1; j < len(result); j++ {
			a2 := result[j]
			if compareUnique(a2, a1[:n]) {
				w := (*r)[:0]
				w = append(w, a1[:n]...)
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
