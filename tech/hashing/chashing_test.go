package hashing_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/phamnam2003/challenges/tech/hashing"
)

func TestConsistentHashing(t *testing.T) {
	t.Parallel()

	servers := []string{"serverA", "serverB", "serverC"}
	ring := hashing.NewHashRing(servers, 100) // 100 replicas mỗi server

	clients := make([]string, 0, 20)
	for range 20 {
		clients = append(clients, uuid.NewString())
	}

	fmt.Println("=== Before adding serverD ===")
	counts := map[string]int{}
	for _, c := range clients {
		server := ring.GetNode(c)
		counts[server]++
		fmt.Printf("Client %-36s -> %s\n", c, server)
	}
	fmt.Println("Distribution:", counts)

	// thêm serverD
	ring.AddNode("serverD")

	fmt.Println("\n=== After adding serverD ===")
	counts = map[string]int{}
	for _, c := range clients {
		server := ring.GetNode(c)
		counts[server]++
		fmt.Printf("Client %-36s -> %s\n", c, server)
	}
	fmt.Println("Distribution:", counts)
}
