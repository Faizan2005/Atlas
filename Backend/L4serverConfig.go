package backend

import (
	"sync"
	"time"
)

type L4ServerOpts struct {
	Address string
	Weight  int
}

type L4BackendServer struct {
	L4ServerOpts
	ConnCount int // For Least Connections
	//AvgLatency    float64 // For Least Response Time
	Alive         bool // Health check status
	LastChecked   time.Time
	StickyClients map[string]bool // Optional: for session stickiness
	Mx            sync.Mutex
}

type L4BackendPool struct {
	Servers []*L4BackendServer
	Mutex   sync.RWMutex
	Index   int // For Round Robin
}

func L4NewServer(Opts L4ServerOpts) *L4BackendServer {
	return &L4BackendServer{
		L4ServerOpts:  Opts,
		Alive:         true,
		StickyClients: make(map[string]bool),
		Mx:            *new(sync.Mutex),
	}
}
