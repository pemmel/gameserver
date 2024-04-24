package auth

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/pemmel/gameserver/protobuf"
	"github.com/pemmel/gameserver/server"
	"google.golang.org/protobuf/proto"
)

const version uint8 = 1

type Config struct {
	Address         string
	TlsCertFile     string
	TlsKeyFile      string
	QueueCapacity   int
	QueueBufferSize int
	NbWorkers       int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
}

// RunAuthServer starts the authentication server with the provided configuration.
// It loads the TLS certificate and key pair from the specified files, creates a TLS listener
// on the given address, and spawns worker goroutines to handle incoming connections.
//
// Parameters:
//
//	c (Config): The configuration for the authentication server.
//
// Returns:
//
//	error: An error if any occurred during server setup, otherwise nil.
func RunAuthServer(c Config) error {
	cert, err := tls.LoadX509KeyPair(c.TlsCertFile, c.TlsKeyFile)
	if err != nil {
		return err
	}

	server, err := tls.Listen("tcp4", c.Address, &tls.Config{
		ClientAuth:   tls.NoClientCert,
		Certificates: []tls.Certificate{cert},
	})
	if err != nil {
		return err
	}

	chc := make(chan net.Conn, c.QueueCapacity)
	go director(server, chc)
	for i := 0; i < c.NbWorkers; i++ {
		go handler(chc, c)
	}

	return nil
}

// director accepts incoming connections on the provided server listener
// and forwards them to the specified channel for handling.
//
// Parameters:
//
//	server (net.Listener): The listener on which to accept incoming connections.
//	chc (chan net.Conn): The channel to which accepted connections are forwarded.
func director(server net.Listener, chc chan net.Conn) {
	for {
		conn, err := server.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		chc <- conn
	}
}

// handler processes incoming connections received from the specified channel.
// It performs authentication and session management for each connection.
//
// Parameters:
//
//	chc (chan net.Conn): The channel from which connections are received.
//	bufferSize (int): The size of the buffer used for reading data from connections.
//	readTimeout (time.Duration): The timeout for read operations on connections.
func handler(chc chan net.Conn, c Config) {
	for conn := range chc {
		var err error

		// Set write deadline to minimize congestion
		err = conn.SetWriteDeadline(time.Now().Add(c.WriteTimeout))
		if err != nil {
			close(conn, responseUnknown)
			continue
		}

		// Set read deadline to minimize congestion.
		err = conn.SetReadDeadline(time.Now().Add(c.ReadTimeout))
		if err != nil {
			close(conn, responseInternalError)
			continue
		}

		// Read auth login request from the client.
		b := make([]byte, c.QueueBufferSize)
		n, err := conn.Read(b)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				close(conn, responseLoginTimeout)
				continue
			} else {
				close(conn, responseInternalError)
				continue
			}
		}

		// Verify client auth token and claims.
		s := string(b[:n])
		c := VerifyAuthClaims(s)
		if c == nil {
			close(conn, responseInvalidToken)
			continue
		}

		// Ensure the claims match with this server.
		ok := VerifyServerMatch(c.IdProvider, c.AppId)
		if !ok {
			close(conn, responseInvalidServer)
			continue
		}

		// Ensure no session is associated with this user
		if server.SharedSession().GetFromUid(c.Uid) != nil {
			close(conn, responseLoginConflict)
			continue
		}

		// Create a new session for this user id
		new := server.NewSessionV1
		session := server.SharedSession().NewSession(new, c.Uid)
		if session == nil {
			close(conn, responseInternalError)
			continue
		}

		// Write response for login success
		r := &protobuf.AuthResponseLoginSuccess{
			Sidx:      session.Sidx,
			Aes256Key: session.SharedKey[:],
		}

		p, err := proto.Marshal(r)
		if err != nil {
			close(conn, responseInternalError)
			continue
		}

		b = b[:0]
		b = append(b, version)
		b = append(b, responseLoginSuccess)
		b = append(b, p...)
		conn.Write(b)
		conn.Close()
	}
}

// Returning a response on failed condition optional and non-crucial.
// Just make sure to close the connection to free up resources.
func close(c net.Conn, r response) {
	if r != responseUnknown {
		b := [2]byte{version, r}
		c.Write(b[:])
	}
	c.Close()
}
