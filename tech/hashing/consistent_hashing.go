// Package hashing implements consistent hashing for distributing keys across a set of nodes (SHARDING).
package hashing

import (
	"fmt"
	"hash/fnv"
	"sort"
)

// ServerNode represents a server node in the consistent hashing ring.
type ServerNode struct {
	ID   string
	Hash uint32
}

// HashRing represents the consistent hashing ring.
type HashRing struct {
	nodes    []ServerNode
	replicas int
}

// hashKey computes the FNV-1a hash of a given key.
func hashKey(key string) uint32 {
	hash := fnv.New32a()
	hash.Write([]byte(key))

	return hash.Sum32()
}

// NewHashRing creates a new HashRing with the given server nodes and replicas.
func NewHashRing(servers []string, replicas int) *HashRing {
	ring := &HashRing{replicas: replicas}
	for _, s := range servers {
		ring.AddNode(s)
	}
	return ring
}

// AddNode adds a new server node to the hash ring with virtual nodes.
func (h *HashRing) AddNode(id string) {
	for i := 0; i < h.replicas; i++ {
		vnodeID := fmt.Sprintf("%s#%d", id, i)
		n := ServerNode{ID: id, Hash: hashKey(vnodeID)}
		h.nodes = append(h.nodes, n)
	}
	sort.Slice(h.nodes, func(i, j int) bool {
		return h.nodes[i].Hash < h.nodes[j].Hash
	})
}

// GetNode finds the server node responsible for a given key.
func (h *HashRing) GetNode(key string) string {
	if len(h.nodes) == 0 {
		return ""
	}
	kh := hashKey(key)

	// binary search return nearest node by clockwise direction
	idx := sort.Search(len(h.nodes), func(i int) bool {
		return h.nodes[i].Hash >= kh
	})
	if idx == len(h.nodes) {
		idx = 0
	}
	return h.nodes[idx].ID
}
