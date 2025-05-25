package network

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	algorithm "github.com/Faizan2005/Balancer"
)

func (lb *LBProperties) HandleHTTP(peekReader *bufio.Reader, conn net.Conn) {
	defer conn.Close()
	startTime := time.Now()
	log.Println("[HTTP_HANDLER] New HTTP connection received.")

	// Capture buffered data and prepend it to the rest of the stream
	peeked, err := peekReader.Peek(peekReader.Buffered())
	if err != nil {
		log.Printf("[HTTP_HANDLER] Error peeking buffered data: %v", err)
		return
	}
	reader := io.MultiReader(bytes.NewReader(peeked), peekReader)

	// Parse HTTP request for routing
	req, err := http.ReadRequest(bufio.NewReader(reader))
	if err != nil {
		log.Printf("[HTTP_HANDLER] Error parsing HTTP request: %v", err)
		return
	}

	path := req.URL.Path
	urlType := ClassifyURLRequest(path)
	pool := lb.L7LBProperties.L7Pools[urlType]
	if pool == nil {
		log.Printf("[HTTP_HANDLER] No server pool found for URL type: %s", urlType)
		return
	}

	l7Adapter := algorithm.L7PoolAdapter{pool}
	algoName := algorithm.SelectAlgoL7(&l7Adapter)
	if algoName == "" {
		log.Println("[HTTP_HANDLER] No algorithm selected for L7 request")
		return
	}

	server := algorithm.ApplyAlgo(&l7Adapter, algoName, lb.AlgorithmsMap)
	if server == nil {
		log.Println("[HTTP_HANDLER] No server returned by algorithm")
		return
	}

	log.Printf("[HTTP_HANDLER] Selected backend: %s", server.GetAddress())

	// Update server connection count
	server.Lock()
	server.SetConnCount(server.GetConnCount() + 1)
	server.Unlock()

	backendConn, err := net.Dial("tcp", server.GetAddress())
	if err != nil {
		log.Printf("[HTTP_HANDLER] Failed to connect to backend %s: %v", server.GetAddress(), err)
		return
	}
	defer backendConn.Close()

	// Reuse reader for forwarding (request already parsed)
	go func() {
		_, err := io.Copy(backendConn, reader)
		if err != nil {
			log.Printf("[HTTP_HANDLER] Error copying client → backend: %v", err)
		}
	}()

	_, err = io.Copy(conn, backendConn)
	if err != nil {
		log.Printf("[HTTP_HANDLER] Error copying backend → client: %v", err)
	}

	log.Printf("[HTTP_HANDLER] TCP forwarding done for path %s", path)
	log.Printf("[HTTP_HANDLER] Total time taken: %v", time.Since(startTime))

	// Decrement server connection count
	server.Lock()
	server.SetConnCount(server.GetConnCount() - 1)
	server.Unlock()
}

func ClassifyURLRequest(path string) string {
	staticExt := []string{".jpg", ".jpeg", ".png", ".gif", ".css", ".js", ".ico", ".html"}

	for _, s := range staticExt {
		if strings.HasSuffix(path, s) {
			return "static"
		}

	}

	return "dynamic"
}
