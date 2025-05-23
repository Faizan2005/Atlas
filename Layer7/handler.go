package layer7

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net"
	"net/http"
)

func HandleHTTP(data []byte, conn net.Conn) {
	peekReader := bytes.NewReader(data)
	bufReader := bufio.NewReader(io.MultiReader(peekReader, conn))

	req, err := http.ReadRequest(bufReader)
	if err != nil {
		log.Printf("Error reading incoming HTTP request: %v", err)
		return
	}

	host := req.Host                      // Host header
	path := req.URL.Path                  // URL path
	cookie, _ := req.Cookie("session_id") // Cookie by name
	userAgent := req.Header.Get("User-Agent")
	log.Printf("Host header: %s\n URL path: %s\n Cookie: %v\n User-Agent header: %s", host, path, cookie, userAgent)

}
