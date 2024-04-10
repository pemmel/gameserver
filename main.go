package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"
)

var (
	addr      *net.UDPAddr
	nbWorkers int
)

func main() {
	var port int
	var conn *net.UDPConn
	var err error = nil
	var env string

	env = os.Getenv("PORT")
	if port, err = strconv.Atoi(env); err != nil {
		port = 8181
	}
	addr = &net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("0.0.0.0"),
	}
	log.Printf("Server UDP:%v", addr.String())
	if conn, err = net.ListenUDP("udp4", addr); err != nil {
		log.Panicln(err)
	}
	env = os.Getenv("WRITE_BUFFER_SIZE")
	if s, err := strconv.Atoi(env); err == nil {
		if err = conn.SetWriteBuffer(s); err != nil {
			log.Panicln(err)
		}
	}
	env = os.Getenv("READ_BUFFER_SIZE")
	if s, err := strconv.Atoi(env); err == nil {
		if err = conn.SetReadBuffer(s); err != nil {
			log.Panicln(err)
		}
	}

	nbWorkers = runtime.NumCPU()
	runtime.GOMAXPROCS(nbWorkers)
	m := NewMetrics()
	run(conn, m)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGABRT)
	go func() {
		<-c
		m.flush()
		log.Printf("Total Ops: %d", m.opsTotal)
		os.Exit(0)
	}()

	t := time.NewTicker(time.Second)
	for range t.C {
		log.Printf("Ops: %d", m.flush())
	}
}
