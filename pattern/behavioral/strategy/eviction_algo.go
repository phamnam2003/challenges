package main

// EvictionAlgo is the strategy interface for cache eviction algorithms
type EvictionAlgo interface {
	evict(c *Cache)
}
