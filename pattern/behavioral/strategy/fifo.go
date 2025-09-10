package main

import "fmt"

// Fifo is a concrete strategy
type Fifo struct{}

// evict evicts an item from the cache using fifo strategy
func (l *Fifo) evict(c *Cache) {
	fmt.Println("Evicting by fifo strtegy")
}
