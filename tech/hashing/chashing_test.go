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
	ring := hashing.NewHashRing(servers, 256)

	clients := make([]string, 0, 20)
	for range 1_000_000 {
		clients = append(clients, uuid.NewString())
	}

	fmt.Println("=== Before adding serverD ===")
	counts := map[string]int{}
	for _, c := range clients {
		server := ring.GetNode(c)
		counts[server]++
	}
	fmt.Println("Distribution:", counts)

	ring.AddNode("serverD")

	fmt.Println("\n=== After adding serverD ===")
	counts = map[string]int{}
	for _, c := range clients {
		server := ring.GetNode(c)
		counts[server]++
	}
	fmt.Println("Distribution:", counts)
}
