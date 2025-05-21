package balancer

import (
	"log"

	backend "github.com/Faizan2005/Backend"
)

type AlgoProperty struct {
	Name      string
	Priority  int
	Condition func(pool *backend.BackendPool) bool
}

var AlgoRules = []AlgoProperty{
	// {
	//     Name: "ip_hash",
	//     Priority: 1,
	//     Condition: NeedsSessionAffinity,
	// },
	{
		Name:      "least_connection",
		Priority:  1,
		Condition: HasLoadImbalance,
	},
	{
		Name:      "weighted_round_robin",
		Priority:  2,
		Condition: HasUnevenWeights,
	},
	{
		Name:      "round_robin",
		Priority:  3,
		Condition: HasUnevenWeights,
	},
}

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

func SelectAlgo(pool *backend.BackendPool) []string {
	// if HasUnevenWeights(pool) {
	// 	return "weighted_round_robin"
	// }

	// return "round_robin"
	var selected []string

	for _, a := range AlgoRules {
		if a.Condition(pool) {
			selected = append(selected, a.Name)
		}
	}

	return selected
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

func HasLoadImbalance(pool *backend.BackendPool) bool {
	pool.Mutex.Lock()
	defer pool.Mutex.Unlock()

	if len(pool.Servers) < 2 {
		return false
	}

	max, min := pool.Servers[0].ConnCount, pool.Servers[0].ConnCount

	for _, s := range pool.Servers {
		if s.ConnCount > max {
			max = s.ConnCount
		}

		if s.ConnCount < min {
			min = s.ConnCount
		}
	}

	return max-min >= 10
}

func ApplyAlgoChain(pool *backend.BackendPool, algoNames []string, algo map[string]LBStrategy) *backend.BackendServer {
	for _, name := range algoNames {
		strategy, exists := algo[name]
		if !exists {
			log.Printf("Algorithm %s not implemented", name)
			continue
		}

		server := strategy.ImplementAlgo(pool)
		if server != nil {
			return server // You could support deeper chaining too
		}
	}

	return nil
}
