package main

import (
	"fmt"
	"sync"
	"time"

	backend "github.com/Faizan2005/Backend"
	tcp "github.com/Faizan2005/Layer4/Transport"
	L7 "github.com/Faizan2005/Layer7"
)

func main() {
	opts := tcp.TransportOpts{
		ListenAddr: ":3000",
	}

	transport := tcp.NewTCPTransport(opts)

	L4pool := backend.L4BackendPool{
		Servers: backend.MakeL4TestServers(),
		Mutex:   *new(sync.RWMutex),
	}

	staticPoolOpts := backend.L7PoolOpts{
		Name:    "static",
		Servers: backend.MakeL7StaticTestServers(),
	}

	staticPool := backend.NewL7ServerPool(staticPoolOpts)

	dynamicPoolOpts := backend.L7PoolOpts{
		Name:    "dynamic",
		Servers: backend.MakeL7DynamicTestServers(),
	}

	dynamicPool := backend.NewL7ServerPool(dynamicPoolOpts)

	L7pools := map[string]*backend.L7ServerPool{
		"static":  staticPool,
		"dynamic": dynamicPool,
	}

	L7Prop := L7.L7LBProperties{
		L7Pools: L7pools,
	}

	p := tcp.NewLBProperties(*transport, L4pool, &L7Prop)

	if err := p.ListenAndAccept(); err != nil {
		panic(err)
	}

	go ClientServer()

	go func() {
		for {
			time.Sleep(3 * time.Second)
			fmt.Println("=== Backend Server States ===")
			for _, srv := range L4pool.Servers {
				fmt.Printf("Server: %s | ConnCount: %d\n | Weight: %d\n", srv.Address, srv.ConnCount, srv.Weight)
			}
			fmt.Println("=============================")
		}
	}()

	select {}
}
