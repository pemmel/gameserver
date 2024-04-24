package test

import (
	"testing"

	"github.com/pemmel/gameserver/server/game"
)

func TestLList(t *testing.T) {
	// seeds the list with expected value
	llist := game.NewLList[int]()
	for i := 0; i < 1e8; i++ {
		llist.Insert(i)
	}

	// assert the value in the list
	next := llist.Next()
	for i := 0; i < 1e8; i++ {
		if next.Value != i {
			t.FailNow()
		}
		next = next.Next()
	}
}
