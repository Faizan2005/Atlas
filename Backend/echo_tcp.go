package backend

import (
	"fmt"
	"log"
	"math/rand"
	"net"
)

func MakeTestServers() []*BackendServer {
	var servers []*BackendServer

	for i := 1; i <= 5; i++ {
		addr := fmt.Sprintf(":900%d", i)
		opts := ServerOpts{
			Address: addr,
			Weight:  rand.Intn(10) + 1,
		}
		server := NewServer(opts)
		server.testServerListener()
		servers = append(servers, server)
	}

	return servers
}

func (bs *BackendServer) testServerListener() {
	listener, err := net.Listen("tcp", bs.Address)
	if err != nil {
		log.Printf("Error listening from server %s: %v", bs.Address, err)
		return
	}

	log.Printf("Backend server started on %s", bs.Address)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Error establishing connection on %s: %v", bs.Address, err)
				continue
			}

			go func(c net.Conn) {
				defer c.Close()
				buff := make([]byte, 1024)
				n, err := c.Read(buff)
				if err != nil {
					log.Printf("Error reading from Load Balancer: %v", err)
					return
				}

				log.Printf("Received (%d) bytes from Load Balancer on %s", n, bs.Address)

				// Echo back a dummy response
				c.Write([]byte("Hello from backend " + bs.Address + "\n"))
			}(conn)
		}
	}()
}
