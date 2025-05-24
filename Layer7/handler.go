package layer7

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net"
	"net/http"
	"strings"

	backend "github.com/Faizan2005/Backend"
	algorithm "github.com/Faizan2005/Balancer"
)

type L7LBProperties struct {
	L7Pools map[string]*backend.L7ServerPool
}

func NewL7LBProperties(pools map[string]*backend.L7ServerPool) *L7LBProperties {
	return &L7LBProperties{
		L7Pools: pools,
	}
}

func (lb *L7LBProperties) HandleHTTP(data []byte, conn net.Conn) {
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

	urlType := ClassifyURLRequest(path)
	pool := lb.L7Pools[urlType]
	L7PoolAdaptr := algorithm.L7PoolAdapter{pool}

	algoName := algorithm.SelectAlgoL7(&L7PoolAdaptr)

	log.Printf("Selected algo to implement (%s)", algoName)

	server := algorithm.ApplyAlgo(&L7PoolAdaptr, algoName)

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
