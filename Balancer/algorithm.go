package balancer

import (
	"fmt"
	"hash/fnv"
	"log"
	"net"

	backend "github.com/Faizan2005/Backend"
)

type AlgoProperty struct {
	Name      string
	Priority  int
	Condition func(pool *backend.L4BackendPool) bool
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
	ImplementAlgo(pool *backend.L4BackendPool) *backend.L4BackendServer
}

// Implementing RR algo
type AlgoRR struct{}

func (rr *AlgoRR) ImplementAlgo(pool *backend.L4BackendPool) *backend.L4BackendServer {
	pool.Mutex.Lock()
	defer pool.Mutex.Unlock()

	n := len(pool.Servers)
	log.Printf("Round Robin: Starting selection from index %d", pool.Index)

	for i := 0; i < n; i++ {
		index := (pool.Index + i) % n
		if pool.Servers[index].Alive {
			log.Printf("Round Robin: Selected server %s at index %d", pool.Servers[index].Address, index)
			pool.Index = index + 1
			return pool.Servers[index]
		}
	}

	log.Println("Round Robin: No healthy server found")
	return nil // No healthy server found
}

type AlgoWRR struct {
	counter int
}

func (wrr *AlgoWRR) ImplementAlgo(pool *backend.L4BackendPool) *backend.L4BackendServer {
	pool.Mutex.Lock()
	defer pool.Mutex.Unlock()

	total := 0
	for _, s := range pool.Servers {
		if s.Alive {
			total += s.Weight
		}
	}

	if total == 0 {
		log.Println("Weighted Round Robin: No healthy servers available")
		return nil // No healthy servers
	}

	wrr.counter = (wrr.counter + 1) % total
	log.Printf("Weighted Round Robin: Current counter value is %d", wrr.counter)

	sum := 0
	for _, s := range pool.Servers {
		if !s.Alive {
			continue
		}
		sum += s.Weight
		if wrr.counter < sum {
			log.Printf("Weighted Round Robin: Selected server %s with weight %d", s.Address, s.Weight)
			return s
		}
	}

	log.Println("Weighted Round Robin: No server selected")
	return nil
}

type AlgoLeastConn struct{}

func (lc *AlgoLeastConn) ImplementAlgo(pool *backend.L4BackendPool) *backend.L4BackendServer {
	pool.Mutex.Lock()
	defer pool.Mutex.Unlock()

	selected := new(*backend.L4BackendServer)
	minConns := int(^uint(0) >> 1) // Max int

	log.Println("Least Connections: Evaluating servers for least connections")

	for _, s := range pool.Servers {
		s.Mx.Lock()
		cCount := s.ConnCount
		s.Mx.Unlock()

		log.Printf("Least Connections: Server %s has %d connections", s.Address, cCount)

		if selected == nil || cCount < minConns {
			selected = &s
			minConns = cCount
			log.Printf("Least Connections: New selected server %s with %d connections", s.Address, cCount)
		}
	}

	if selected != nil {
		log.Printf("Least Connections: Selected server %s with %d connections", (*selected).Address, minConns)
		return *selected
	}

	log.Println("Least Connections: No server selected")
	return nil
}

func SelectAlgoL4(pool *backend.L4BackendPool) string {
	if HasUnevenWeights(pool) {
		return "weighted_least_connection"
	}
	return "least_connection"
}

func SelectAlgoL7(pool *backend.L4BackendPool) string {
	if HasLoadImbalance(pool) {
		if HasUnevenWeights(pool) {
			return "weighted_least_connection"
		}
		return "least_connection"
	}
	if HasUnevenWeights(pool) {
		return "weighted_round_robin"
	}
	return "round_robin"
}

func HasUnevenWeights(pool *backend.L4BackendPool) bool {
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

func NewWLCountAlgo() LBStrategy {
	return &AlgoWLeastConn{}
}

func IPHash(pool *backend.L4BackendPool, host_ip string) *backend.L4BackendServer {
	pool.Mutex.Lock()
	defer pool.Mutex.Unlock()

	ip, port, err := net.SplitHostPort(host_ip)
	if err != nil {
		fmt.Printf("Error splitting host and port from %s: %v\n", host_ip, err)
		return nil
	}

	fmt.Printf("[IPHash] Client IP: %s, Port: %s\n", ip, port)

	hash := fnv.New32a()
	hash.Write([]byte(ip))
	hashValue := hash.Sum32()
	index := int(hashValue) % len(pool.Servers)

	fmt.Printf("[IPHash] FNV Hash Value: %d, Backend Index: %d\n", hashValue, index)
	fmt.Printf("[IPHash] Selected Backend: %s\n", pool.Servers[index].Address)

	return pool.Servers[index]
}

func HasLoadImbalance(pool *backend.L4BackendPool) bool {
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

	return max-min >= 5
}

func ApplyAlgo(pool *backend.L4BackendPool, algoName string, algo map[string]LBStrategy) *backend.L4BackendServer {
	strategy, exists := algo[algoName]
	if !exists {
		log.Printf("Algorithm %s not implemented", algoName)
	}

	server := strategy.ImplementAlgo(pool)
	if server != nil {
		return server // You could support deeper chaining too
	}

	return nil
}

type AlgoWLeastConn struct{}

func (wlc *AlgoWLeastConn) ImplementAlgo(pool *backend.L4BackendPool) *backend.L4BackendServer {
	pool.Mutex.Lock()
	defer pool.Mutex.Unlock()

	selected := new(*backend.L4BackendServer)
	minScore := int(^uint(0) >> 1) // Max int

	log.Println("Weighted Least Connections: Evaluating servers for least connections")

	for _, s := range pool.Servers {
		s.Mx.Lock()
		score := int(s.ConnCount) / int(s.Weight)
		s.Mx.Unlock()

		log.Printf("Weighted Least Connections: Server %s has %d connections", s.Address, score)

		if selected == nil || score < minScore {
			selected = &s
			minScore = score
			log.Printf("Weighted Least Connections: New selected server %s with %d connections", s.Address, score)
		}
	}

	if selected != nil {
		log.Printf("Weighted Least Connections: Selected server %s with %d score", (*selected).Address, minScore)
		return *selected
	}

	log.Println("Weighted Least Connections: No server selected")
	return nil
}
