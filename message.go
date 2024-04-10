package main

import "net"

type message struct {
	data []byte
	len  int
	addr *net.UDPAddr
}
