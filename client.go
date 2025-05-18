package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

func runClient(id int) {
	conn, err := net.Dial("tcp", "localhost:3000")
	if err != nil {
		fmt.Printf("[Client %d] Connection error: %v\n", id, err)
		return
	}
	defer conn.Close()

	msg := fmt.Sprintf("Hello from client %d", id)
	conn.Write([]byte(msg))

	buf := make([]byte, 1024)
	n, _ := conn.Read(buf)
	fmt.Printf("[Client %d] Received: %s\n", id, string(buf[:n]))
}

func ClientServer() {
	var wg sync.WaitGroup
	clientCount := 20

	for i := 1; i <= clientCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			runClient(id)
		}(i)

		time.Sleep(100 * time.Millisecond) // small delay to stagger connections
	}

	wg.Wait()
}
