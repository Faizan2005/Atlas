package main

import (
	"fmt"
	"sync"
	"time"

	backend "github.com/Faizan2005/Backend"
	tcp "github.com/Faizan2005/Layer4/Transport"
)

func main() {
	opts := tcp.TransportOpts{
		ListenAddr: ":3000",
	}

	transport := tcp.NewTCPTransport(opts)

	pool := backend.L4BackendPool{
		Servers: backend.MakeL4TestServers(),
		Mutex:   *new(sync.RWMutex),
	}

	p := tcp.NewLBProperties(*transport, pool)

	if err := p.ListenAndAccept(); err != nil {
		panic(err)
	}

	go ClientServer()

	go func() {
		for {
			time.Sleep(3 * time.Second)
			fmt.Println("=== Backend Server States ===")
			for _, srv := range pool.Servers {
				fmt.Printf("Server: %s | ConnCount: %d\n | Weight: %d\n", srv.Address, srv.ConnCount, srv.Weight)
			}
			fmt.Println("=============================")
		}
	}()

	select {}
}
