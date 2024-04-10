package main

import (
	"log"
	"net"
)

const (
	QUEUE_CAPACITY int = 1e6
	BUFFER_SIZE    int = 1500
)

func run(conn *net.UDPConn, m *metrics) {
	lcm := make(chan *message, QUEUE_CAPACITY)
	pcm := make(chan *message, QUEUE_CAPACITY)
	for i := 0; i < QUEUE_CAPACITY; i++ {
		lcm <- &message{data: nil}
	}
	go listener(conn, lcm, pcm)
	for i := 0; i < nbWorkers; i++ {
		go processor(lcm, pcm, m)
	}
}

func listener(conn *net.UDPConn, lcm chan *message, pcm chan *message) {
	var err error
	for {
		msg := <-lcm
		if msg.data == nil {
			msg.data = make([]byte, BUFFER_SIZE)
		}
		msg.len, msg.addr, err = conn.ReadFromUDP(msg.data)
		if err != nil {
			log.Println(err)
			continue
		}
		pcm <- msg
	}
}

func processor(lcm chan *message, pcm chan *message, m *metrics) {
	for msg := range pcm {
		m.increment()
		lcm <- msg
	}
}
