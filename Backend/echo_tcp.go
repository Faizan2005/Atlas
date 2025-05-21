package backend

import (
	"fmt"
	"io"
	"log"
	"net"
)

func MakeTestServers() []*BackendServer {
	var servers []*BackendServer

	weights := []int{5, 3, 1} // Highly skewed weights

	for i := 0; i < 3; i++ {
		addr := fmt.Sprintf(":900%d", i)
		opts := ServerOpts{
			Address: addr,
			Weight:  weights[i],
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
				defer func() {
					log.Printf("Closing backend connection with server %s", bs.Address)
					c.Close()
				}()

				buff := make([]byte, 1024)

				for {
					n, err := c.Read(buff)
					if err != nil {
						if err == io.EOF {
							log.Printf("Client closed the connection")
						} else {
							log.Printf("Error reading from Load Balancer: %v", err)
						}
						return
					}

					log.Printf("Received (%d) bytes from Load Balancer on %s", n, bs.Address)

					msg := fmt.Sprintf("Hello from backend %s server weight: %d\n", bs.Address, bs.Weight)
					_, err = c.Write([]byte(msg))
					if err != nil {
						log.Printf("Error writing to Load Balancer: %v", err)
						return
					}
				}
			}(conn)
		}
	}()
}
