package balancer

import (
	pool "github.com/Faizan2005/Backend"
)

// Interface for selecting lb algorithm for different situations
type LBStrategy interface {
	ImplementAlgo(pool *pool.BackendPool) *pool.BackendServer
}

// Implementing RR algo
type AlgoRR struct{}

func (rr *AlgoRR) ImplementAlgo(pool *pool.BackendPool) *pool.BackendServer {
	pool.Mutex.Lock()
	defer pool.Mutex.Unlock()

	server := pool.Servers[pool.Index%len(pool.Servers)]
	pool.Index++

	return server
}

// Implementing Weighted RR algo
type AlgoWRR struct {
	counter int
}

func (wrr *AlgoWRR) ImplementAlgo(pool *pool.BackendPool) *pool.BackendServer {
	pool.Mutex.Lock()
	defer pool.Mutex.Unlock()

	total := 0
	for _, s := range pool.Servers {
		total += s.Weight
	}

	wrr.counter = (wrr.counter + 1) % total

	sum := 0
	for _, s := range pool.Servers {
		sum += s.Weight
		if wrr.counter < sum {
			return s
		}
	}
	return nil

}

func SelectAlgo(pool *pool.BackendPool) string {
	if HasUnevenWeights(pool) {
		return "weighted_round_robin"
	}

	return "round_robin"
}

func HasUnevenWeights(pool *pool.BackendPool) bool {
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
