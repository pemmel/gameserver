package main

import (
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	flushInterval = 1 * time.Second
	UDPPacketSize = 1500
)

var (
	addr        string
	ops         uint64 = 0
	opsTotal    uint64 = 0
	flushTicker *time.Ticker
	nbWorkers   int
	loading     = true
	conn        net.Conn
)

func main() {
	var err error

	addr = os.Getenv("ADDRESS")
	conn, err = net.Dial("udp4", addr)
	if err != nil {
		log.Panicln(err)
		os.Exit(1)
	}

	nbWorkers = runtime.NumCPU()
	runtime.GOMAXPROCS(nbWorkers)
	load(nbWorkers)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGABRT)
	go func() {
		<-c
		loading = false
		log.Printf("Total Ops: %d", opsTotal)
		conn.Close()
		os.Exit(0)
	}()

	flushTicker = time.NewTicker(flushInterval)
	for range flushTicker.C {
		prevOps := atomic.SwapUint64(&ops, 0)
		atomic.AddUint64(&opsTotal, prevOps)
		log.Printf("Ops: %d", prevOps)
	}
}

func load(maxWorkers int) {
	for i := 0; i < maxWorkers; i++ {
		go func() {
			for loading {
				_, err := conn.Write([]byte("hello"))
				if err != nil {
					if errors.Is(err, net.ErrClosed) {
						conn, _ = net.Dial("udp4", addr)
					}
					time.Sleep(time.Second)
					continue
				}
				atomic.AddUint64(&ops, 1)
			}
		}()
	}
}
