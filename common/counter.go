package common

import (
	"fmt"
	"sync/atomic"
	"time"
)

var (
	counters []*Counter
)

func init() {
	counters = make([]*Counter, 0)
}

func CounterSize() int {
	return len(counters)
}

func RunCounterLogging() {
	t := time.NewTicker(time.Second)
	for range t.C {
		for _, c := range counters {
			fmt.Printf("[%s]: %d\n", c.tag, c.Flush())
		}
	}
}

type Counter struct {
	tag   string
	total uint64
	curr  uint64
}

func RegisterNewCounter(tag string) *Counter {
	m := &Counter{
		tag:   tag,
		curr:  0,
		total: 0,
	}
	counters = append(counters, m)
	return m
}

func (m *Counter) Increment() {
	atomic.AddUint64(&m.curr, 1)
}

func (m *Counter) Flush() uint64 {
	old := atomic.SwapUint64(&m.curr, 0)
	atomic.AddUint64(&m.total, old)
	return old
}

func (m *Counter) Total() uint64 {
	return m.total
}
