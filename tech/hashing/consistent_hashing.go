// Package hashing implements consistent hashing for distributing keys across a set of nodes (SHARDING).
package hashing

import (
	"crypto/sha1"
	"sort"
)

// ServerNode represents a server node in the consistent hashing ring.
type ServerNode struct {
	ID   string
	Hash uint32

	// can be load instance the DB connection
}

// HashRing represents the consistent hashing ring.
type HashRing struct {
	nodes []ServerNode
}

// hashKey generates a hash for a given key using SHA-1 and returns the first 4 bytes as a uint32.
func hashKey(key string) uint32 {
	// make SHA-1 hasher, result Sum() is 20 bytes
	h := sha1.New()
	h.Write([]byte(key))
	sum := h.Sum(nil)

	return (uint32(sum[0])<<24 | uint32(sum[1])<<16 | uint32(sum[2])<<8 | uint32(sum[3]))
}

// AddNode adds a new server node to the hash ring.
func (h *HashRing) AddNode(id string) {
	n := ServerNode{ID: id, Hash: hashKey(id)}
	h.nodes = append(h.nodes, n)
	sort.Slice(h.nodes, func(i, j int) bool {
		return h.nodes[i].Hash < h.nodes[j].Hash
	})
}

// NewHashRing creates a new HashRing with the given server nodes.
func NewHashRing(servers []string) *HashRing {
	ring := &HashRing{}
	for _, s := range servers {
		ring.AddNode(s)
	}
	return ring
}

func (h *HashRing) GetNode(key string) string {
	if len(h.nodes) == 0 {
		return ""
	}
	kh := hashKey(key)

	// binary search return nearest node by the clockwise direction
	idx := sort.Search(len(h.nodes), func(i int) bool {
		return h.nodes[i].Hash >= kh
	})
	if idx == len(h.nodes) {
		idx = 0
	}
	return h.nodes[idx].ID
}
