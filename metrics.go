package main

import "sync/atomic"

type metrics struct {
	opsTotal uint64
	ops      uint64
}

func NewMetrics() *metrics {
	return &metrics{
		ops:      0,
		opsTotal: 0,
	}
}

func (m *metrics) increment() {
	atomic.AddUint64(&m.ops, 1)
}

func (m *metrics) flush() uint64 {
	old := atomic.SwapUint64(&m.ops, 0)
	atomic.AddUint64(&m.opsTotal, old)
	return old
}
