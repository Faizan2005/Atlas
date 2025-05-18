package layer4

import (
	"io"
	"log"
	"net"

	backend "github.com/Faizan2005/Backend"
)

type TransportOpts struct {
	ListenAddr string
}

type TCPTransport struct {
	TransportOpts
	Listener net.Listener
}

type LBProperties struct {
	Transport  *TCPTransport
	ServerPool *backend.BackendPool
}

func NewLBProperties(Transport TCPTransport, Pool backend.BackendPool) *LBProperties {
	return &LBProperties{
		Transport:  &Transport,
		ServerPool: &Pool,
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

	defer func() {
		log.Printf("Closing connection with client %s", conn.RemoteAddr())
		conn.Close()
	}()

	server1 := p.ServerPool.Servers[0]
	if server1 == nil {
		log.Println("No servers available")
		return
	}

	backendConn, err := net.Dial("tcp", server1.Address)
	if err != nil {
		log.Printf("Failed to dial backend: %v", err)
		return
	}

	go io.Copy(backendConn, conn) // client → server
	io.Copy(conn, backendConn)    // server → client
	log.Print("echoed msg from server to client")

	defer func() {
		log.Printf("Closing backend connection with server %s", backendConn.RemoteAddr())
		backendConn.Close()
	}()

	// log.Print("Starting reading loop...")
	// for {
	// 	buf := make([]byte, 1024)
	// 	n, err := conn.Read(buf)
	// 	if err != nil {
	// 		log.Printf("Error reading from the connection %v", err)
	// 		return
	// 	}

	// 	log.Printf("Recieved (%d) bytes from the peer", n)
	// }

}
