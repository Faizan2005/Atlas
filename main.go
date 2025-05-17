package main

import (
	tcp "github.com/Faizan2005/Layer4/Transport"
)

func main() {
	opts := tcp.TransportOpts{
		ListenAddr: ":3000",
	}

	transport := tcp.NewTCPTransport(opts)
	transport.ListenAndAccept()

	select {}
}
