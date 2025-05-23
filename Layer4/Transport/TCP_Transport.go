package layer4

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	backend "github.com/Faizan2005/Backend"
	algorithm "github.com/Faizan2005/Balancer"
	L7 "github.com/Faizan2005/Layer7"
)

type TransportOpts struct {
	ListenAddr string
}

type TCPTransport struct {
	TransportOpts
	Listener net.Listener
}

type LBProperties struct {
	Transport     *TCPTransport
	ServerPool    *backend.BackendPool
	AlgorithmsMap map[string]algorithm.LBStrategy
}

func NewLBProperties(Transport TCPTransport, Pool backend.BackendPool) *LBProperties {
	algoMap := map[string]algorithm.LBStrategy{
		"round_robin":               algorithm.NewRRAlgo(),
		"weighted_round_robin":      algorithm.NewWRRAlgo(),
		"least_connection":          algorithm.NewLCountAlgo(),
		"weighted_least_connection": algorithm.NewWLCountAlgo(),
	}
	return &LBProperties{
		Transport:     &Transport,
		ServerPool:    &Pool,
		AlgorithmsMap: algoMap,
	}
}

// type TCPPeer struct {
// 	Conn net.Conn
// }

func NewTCPTransport(opts TransportOpts) *TCPTransport {
	return &TCPTransport{
		TransportOpts: opts,
	}
}

// func NewTCPPeer(conn net.Conn) *TCPPeer {
// 	return &TCPPeer{
// 		Conn: conn,
// 	}
// }

func (p *LBProperties) ListenAndAccept() error {
	var err error

	p.Transport.Listener, err = net.Listen("tcp", p.Transport.ListenAddr)
	if err != nil {
		log.Printf("Failed to listen on %s: %v", p.Transport.ListenAddr, err)
		return err
	}

	go p.loopAndAccept()

	return nil
}

func (p *LBProperties) loopAndAccept() {
	for {
		conn, err := p.Transport.Listener.Accept()
		if err != nil {
			log.Printf("Failed to establish connection with %s: %v", p.Transport.ListenAddr, err)
			return
		}

		go p.handleConn(conn)
	}
}

func (p *LBProperties) handleConn(conn net.Conn) {
	//	peer := NewTCPPeer(conn)
	log.Printf("Connection established with %s", conn.RemoteAddr())

	reader := bufio.NewReader(conn)
	data, err := reader.Peek(16)
	if err != nil {
		log.Println("Error peeking:", err)
	}

	go func() {
		if isHTTP(data[:]) {
			L7.HandleHTTP(data, conn)
		}
	}()

	defer func() {
		log.Printf("Closing connection with client %s", conn.RemoteAddr())
		conn.Close()
	}()

	// go func() {
	// 	for {
	// 		p.ServerPool.HealthChecker()
	// 		time.Sleep(5 * time.Second)
	// 	}
	// }()

	algoName := algorithm.SelectAlgoL4(p.ServerPool)

	log.Printf("Selected algo to implement (%s)", algoName)
	// algo := p.AlgorithmsMap[algoName]
	// server := algo.ImplementAlgo(p.ServerPool)
	server := algorithm.ApplyAlgo(p.ServerPool, algoName, p.AlgorithmsMap)

	server.Mx.Lock()
	server.ConnCount++
	server.Mx.Unlock()

	backendConn, err := net.Dial("tcp", server.Address)
	if err != nil {
		log.Printf("Failed to dial backend: %v", err)
		return
	}

	go io.Copy(backendConn, conn) // client → server
	io.Copy(conn, backendConn)    // server → client
	log.Print("echoed msg from server to client")

	server.Mx.Lock()
	server.ConnCount--
	server.Mx.Unlock()

	defer func() {
		log.Printf("Closing backend connection with server %s", backendConn.RemoteAddr())
		backendConn.Close()
	}()
}

func isHTTP(data []byte) bool {
	methods := []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"}

	for _, m := range methods {
		if strings.HasPrefix(string(data), m+" ") {
			fmt.Printf("Detected HTTP method: %s\n", m)
			return true
		}
	}

	fmt.Println("Not an HTTP method")
	return false
}
