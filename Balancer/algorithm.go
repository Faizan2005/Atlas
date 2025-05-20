package balancer

import (
	backend "github.com/Faizan2005/Backend"
)

// Interface for selecting lb algorithm for different situations
type LBStrategy interface {
	ImplementAlgo(pool *backend.BackendPool) *backend.BackendServer
}

// Implementing RR algo
type AlgoRR struct{}

func (rr *AlgoRR) ImplementAlgo(pool *backend.BackendPool) *backend.BackendServer {
	pool.Mutex.Lock()
	defer pool.Mutex.Unlock()

	n := len(pool.Servers)
	for i := 0; i < n; i++ {
		index := (pool.Index + i) % n
		if pool.Servers[index].Alive {
			pool.Index = index + 1
			return pool.Servers[index]
		}
	}
	return nil // No healthy server found
}

// Implementing Weighted RR algo
type AlgoWRR struct {
	counter int
}

func (wrr *AlgoWRR) ImplementAlgo(pool *backend.BackendPool) *backend.BackendServer {
	pool.Mutex.Lock()
	defer pool.Mutex.Unlock()

	total := 0
	for _, s := range pool.Servers {
		if s.Alive {
			total += s.Weight
		}
	}

	if total == 0 {
		return nil // No healthy servers
	}

	wrr.counter = (wrr.counter + 1) % total

	sum := 0
	for _, s := range pool.Servers {
		if !s.Alive {
			continue
		}
		sum += s.Weight
		if wrr.counter < sum {
			return s
		}
	}

	return nil
}

type AlgoLeastConn struct{}

func (lc *AlgoLeastConn) ImplementAlgo(pool *backend.BackendPool) *backend.BackendServer {
	pool.Mutex.Lock()
	defer pool.Mutex.Unlock()

	selected := new(*backend.BackendServer)
	minConns := int(^uint(0) >> 1) // Max int

	for _, s := range pool.Servers {
		s.Mx.Lock()
		cCount := s.ConnCount
		s.Mx.Unlock()

		if selected == nil || cCount < minConns {
			selected = &s
			minConns = cCount
		}

	}

	return *selected
}

func SelectAlgo(pool *backend.BackendPool) string {
	if HasUnevenWeights(pool) {
		return "weighted_round_robin"
	}

	return "round_robin"
}

func HasUnevenWeights(pool *backend.BackendPool) bool {
	pool.Mutex.RLock()
	defer pool.Mutex.RUnlock()

	if len(pool.Servers) == 0 {
		return false
	}

	ref := pool.Servers[0].Weight
	for _, s := range pool.Servers[1:] {
		if s.Weight != ref {
			return true
		}
	}
	return false
}

func NewRRAlgo() LBStrategy {
	return &AlgoRR{}
}

func NewWRRAlgo() LBStrategy {
	return &AlgoWRR{
		counter: 0}
}

func NewLCountAlgo() LBStrategy {
	return &AlgoLeastConn{}
}
