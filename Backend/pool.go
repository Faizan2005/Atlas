package backend

import (
	"sync"
	"time"
)

type ServerOpts struct {
	Address string
	Weight  int
}

type BackendServer struct {
	ServerOpts
	ConnCount int // For Least Connections
	//AvgLatency    float64 // For Least Response Time
	Alive         bool // Health check status
	LastChecked   time.Time
	StickyClients map[string]bool // Optional: for session stickiness
	Mx            sync.Mutex
}

type BackendPool struct {
	Servers []*BackendServer
	Mutex   sync.RWMutex
	Index   int // For Round Robin
}

func NewServer(Opts ServerOpts) *BackendServer {
	return &BackendServer{
		ServerOpts:    Opts,
		Alive:         true,
		StickyClients: make(map[string]bool),
		Mx:            *new(sync.Mutex),
	}
}
