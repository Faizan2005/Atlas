package backend

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func MakeL7StaticTestServers() []*L7BackendServer {
	var servers []*L7BackendServer

	weights := []int{5, 3, 1} // Highly skewed weights

	for i := 0; i < 3; i++ {
		addr := fmt.Sprintf(":800%d", i)
		opts := L7ServerOpts{
			Address: addr,
			Weight:  weights[i],
		}
		server := NewL7Server(opts)
		server.testStaticServerListener()
		servers = append(servers, server)
	}

	return servers
}

func (s *L7BackendServer) testStaticServerListener() {
	fs := http.FileServer(http.Dir("./static"))
	mux := http.NewServeMux()
	mux.Handle("/", fs)

	server := &http.Server{
		Addr:    s.Address,
		Handler: mux,
	}

	log.Printf("Static server running on %s", s.Address)
	err := server.ListenAndServe()
	if err != nil {
		log.Printf("Error listening on port (%s): %v", s.Address, err)
	}
}

func MakeL7DynamicTestServers() []*L7BackendServer {
	var servers []*L7BackendServer

	weights := []int{5, 3, 1} // Highly skewed weights

	for i := 0; i < 3; i++ {
		addr := fmt.Sprintf(":801%d", i)
		opts := L7ServerOpts{
			Address: addr,
			Weight:  weights[i],
		}
		server := NewL7Server(opts)
		server.testDynamicServerListener()
		servers = append(servers, server)
	}

	return servers
}

func (s *L7BackendServer) testDynamicServerListener() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", dynamicHandlerFunc)

	server := &http.Server{
		Addr:    s.Address,
		Handler: mux,
	}

	log.Printf("Dynamic server running on %s", s.Address)
	err := server.ListenAndServe()
	if err != nil {
		log.Printf("Error listening on port (%s): %v", s.Address, err)
	}
}

func dynamicHandlerFunc(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"path":   r.URL.Path,
		"method": r.Method,
		"msg":    "Handled by API server",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
