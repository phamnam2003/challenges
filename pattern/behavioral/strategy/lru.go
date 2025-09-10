package main

import "fmt"

// Lru is a concrete strategy
type Lru struct{}

// evict evicts an item from the cache using the LRU strategy
func (l *Lru) evict(c *Cache) {
	fmt.Println("Evicting by lru strtegy")
}
