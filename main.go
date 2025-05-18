package main

import (
	"sync"

	backend "github.com/Faizan2005/Backend"
	tcp "github.com/Faizan2005/Layer4/Transport"
)

func main() {
	opts := tcp.TransportOpts{
		ListenAddr: ":3000",
	}

	transport := tcp.NewTCPTransport(opts)

	pool := backend.BackendPool{
		Servers: backend.MakeTestServers(),
		Mutex:   *new(sync.RWMutex),
	}

	p := tcp.NewLBProperties(*transport, pool)

	if err := p.ListenAndAccept(); err != nil {
		panic(err)
	}

	go ClientServer()

	select {}

}
