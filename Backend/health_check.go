package backend

import (
	"log"
	"net"
	"time"
)

func (pool *BackendPool) HealthChecker() {
	for {

		pool.Mutex.Lock()

		for _, s := range pool.Servers {
			conn, err := net.DialTimeout("tcp", s.Address, 2*time.Second)
			if err != nil {
				s.Alive = false
				log.Printf("[HealthCheck] %s is down", s.Address)
			} else {
				s.Alive = true
				log.Printf("[HealthCheck] %s is up and running", s.Address)
				conn.Close()
			}

		}

		pool.Mutex.Unlock()
		time.Sleep(3 * time.Second)
	}
}
