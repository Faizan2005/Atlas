package balancer

import (
	pool "github.com/Faizan2005/Backend"
)

// Interface for selecting lb algorithm for different situations
type LBStrategy interface {
	AlgoSelector(pool *pool.BackendPool) *pool.BackendServer
}

// Implementing RR algo
type algoRR struct{}

func (rr *algoRR) AlgoSelector(pool *pool.BackendPool) *pool.BackendServer {
	pool.Mutex.Lock()
	defer pool.Mutex.Unlock()

	server := pool.Servers[pool.Index%len(pool.Servers)]
	pool.Index++

	return server
}
