package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/pemmel/gameserver/common"
	"github.com/pemmel/gameserver/server/auth"
	"github.com/pemmel/gameserver/server/game"
)

func main() {
	var err error = nil

	ncpu := runtime.NumCPU()
	runtime.GOMAXPROCS(ncpu)

	gamePort, _ := strconv.Atoi(os.Getenv("GAME_PORT"))
	authAddress := fmt.Sprintf("0.0.0.0:%s", os.Getenv("AUTH_PORT"))

	err = game.RunGameServer(game.Config{
		Address: &net.UDPAddr{
			IP:   net.ParseIP("0.0.0.0"),
			Port: gamePort,
		},
		QueueCapacity:   1e6,
		QueueBufferSize: 1500,
		NbWorkers:       ncpu,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	err = auth.RunAuthServer(auth.Config{
		Address:         authAddress,
		TlsCertFile:     "ca-cert.pem",
		TlsKeyFile:      "ca-key.pem",
		QueueCapacity:   100,
		QueueBufferSize: 1024,
		NbWorkers:       ncpu,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    2 * time.Second,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGABRT)
	go func() {
		<-c
		os.Exit(0)
	}()

	common.RunCounterLogging()
}
