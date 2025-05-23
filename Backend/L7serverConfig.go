package backend

import (
	"sync"
	"time"
)

type L7ServerOpts struct {
	Address string
	Weight  int
}

type L7BackendServer struct {
	L7ServerOpts
	ReqCount int // For Least Connections
	//AvgLatency    float64 // For Least Response Time
	Alive         bool // Health check status
	LastChecked   time.Time
	StickyClients map[string]bool // Optional: for session stickiness
	Mx            sync.Mutex
}

type L7ServerPool struct {
	Name    string
	Servers []*L7BackendServer
	Mutex   sync.RWMutex
	Index   int // For Round Robin
}

func L7NewServer(Opts L7ServerOpts) *L7BackendServer {
	return &L7BackendServer{
		L7ServerOpts:  Opts,
		Alive:         true,
		StickyClients: make(map[string]bool),
		Mx:            *new(sync.Mutex),
	}
}
