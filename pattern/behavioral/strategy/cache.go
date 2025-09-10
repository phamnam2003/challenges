package main

// Cache provides a simple key-value store with eviction strategies
type Cache struct {
	storage      map[string]string
	evictionAlgo EvictionAlgo
	capacity     int
	maxCapacity  int
}

// initCache initializes the cache with a given eviction algorithm
func initCache(e EvictionAlgo) *Cache {
	storage := make(map[string]string)
	return &Cache{
		storage:      storage,
		evictionAlgo: e,
		capacity:     0,
		maxCapacity:  2,
	}
}

// setEvictionAlgo sets the eviction algorithm for the cache
func (c *Cache) setEvictionAlgo(e EvictionAlgo) {
	c.evictionAlgo = e
}

// add adds a key-value pair to the cache, evicting if necessary
func (c *Cache) add(key, value string) {
	if c.capacity == c.maxCapacity {
		c.evict()
	}
	c.capacity++
	c.storage[key] = value
}

// get retrieves a value by key from the cache
func (c *Cache) get(key string) {
	delete(c.storage, key)
}

// evict removes an item from the cache based on the eviction strategy
func (c *Cache) evict() {
	c.evictionAlgo.evict(c)
	c.capacity--
}
