package Backend

import (
	"sync"
	"time"
)

type BackendServer struct {
	Address       string  // IP:Port
	Weight        int     // For Weighted Round Robin
	ConnCount     int     // For Least Connections
	AvgLatency    float64 // For Least Response Time
	Alive         bool    // Health check status
	LastChecked   time.Time
	StickyClients map[string]bool // Optional: for session stickiness
	mu            sync.Mutex
}

type BackendPool struct {
	Backends []*BackendServer
	Mutex    sync.RWMutex
	Index    int // For Round Robin
}
