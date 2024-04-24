package game

import (
	"net"

	"github.com/pemmel/gameserver/server"
)

type Config struct {
	Address         *net.UDPAddr
	QueueCapacity   int
	QueueBufferSize int
	NbWorkers       int
}

// RunGameServer starts the game server with the provided configuration.
// It listens for UDP connections on the given address, spawns a listener
// goroutine to handle incoming packets, and spawns worker goroutines
// to process the packets concurrently.
//
// Parameters:
//   - c (Config): The configuration for the game server.
//
// Returns:
//   - error: An error if any occurred during server setup, otherwise nil.
//
// RunGameServer initializes and starts the game server based on the provided configuration 'c'.
// It establishes a UDP listener on the specified address, spawns a listener goroutine to handle
// incoming packets, and creates worker goroutines to process the packets concurrently. If any errors
// occur during server setup, an error is returned; otherwise, nil is returned to indicate successful
// server initialization.
func RunGameServer(c Config) error {
	server, err := net.ListenUDP("udp4", c.Address)
	if err != nil {
		return err
	}

	chp := make(chan packet, c.QueueCapacity)
	go listener(server, chp, c.QueueBufferSize)
	for i := 0; i < c.NbWorkers; i++ {
		go handler(chp)
	}

	go matchmaking()

	return nil
}

// listener reads incoming UDP packets from the provided UDP connection
// and forwards them to the specified channel for processing.
//
// Parameters:
//   - server (*net.UDPConn): The UDP connection to read packets from.
//   - chp (chan *packet): The channel to which incoming packets are forwarded.
//   - bufferSize (int): The size of the buffer used for reading packets.
//
// This function continuously reads incoming UDP packets from the UDP connection 'server'.
// Each packet is then forwarded to the channel 'chp' for further processing. The size of the
// buffer used for reading packets is determined by the 'bufferSize' parameter.
func listener(server *net.UDPConn, chp chan packet, bufferSize int) {
	b := make([]byte, bufferSize)
	for {
		n, addr, err := server.ReadFromUDP(b)
		if err != nil || packetMeaningful(b[:n]) {
			continue
		}

		data := make([]byte, n)
		copy(data, b[:n])
		chp <- packet{addr, data}
	}
}

// handler processes incoming packets received from the specified channel.
// It retrieves session data associated with the packet's source IP address
// and handles the packet accordingly. Only packets from registered sessions are handled.
//
// Parameters:
//   - chp (chan *packet): The channel from which packets are received.
//
// The function processes packets received from the channel 'chp'. It fetches session data
// based on the source IP address of each packet and handles them appropriately. Only packets
// from sessions that are already registered are processed.
func handler(chp chan packet) {
	var gpb [200]byte
	for p := range chp {
		h := p.verify(server.SharedSession(), gpb[:])
		if h != nil {
			handle(h)
		}
	}
}
