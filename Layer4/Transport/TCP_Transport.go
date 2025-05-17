package layer4

import (
	"log"
	"net"
)

type TransportOpts struct {
	ListenAddr string
}

type TCPTransport struct {
	TransportOpts
	Listener net.Listener
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

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.Listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		log.Printf("Failed to listen on %s: %v", t.Listener, err)
		return err
	}

	go t.loopAndAccept()

	return nil
}

func (t *TCPTransport) loopAndAccept() {
	for {
		conn, err := t.Listener.Accept()
		if err != nil {
			log.Printf("Failed to establish connection with %s: %v", t.ListenAddr, err)
			return
		}

		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	//	peer := NewTCPPeer(conn)
	log.Printf("Connection established with %s", conn.RemoteAddr())

	defer func() {
		log.Printf("Closing connection with %s", conn.RemoteAddr())
		conn.Close()
	}()

	log.Print("Starting reading loop...")
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("Error reading from the connection %v", err)
			return
		}

		log.Printf("Recieved (%d) bytes from the peer", n)
	}

}
