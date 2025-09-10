package main

import "fmt"

// Lfu is a concrete strategy
type Lfu struct{}

// evict evicts an item from the cache
func (l *Lfu) evict(c *Cache) {
	fmt.Println("Evicting by lfu strtegy")
}
