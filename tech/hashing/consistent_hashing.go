// Package hashing implements consistent hashing for distributing keys across a set of nodes (SHARDING).
package hashing

import (
	"hash/fnv"
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

// hashKey computes the FNV-1a hash of a given key [can use another hash func, like sha-1, ...].
func hashKey(key string) uint32 {
	hash := fnv.New32a() // init fnv hasher 32-bit
	hash.Write([]byte(key))

	return hash.Sum32()
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
