// Package lrucache similar to linkedlist provides a generic LRU cache implementation.
package lrucache

type Node struct {
	Prev     *Node
	Key, Val int
	Next     *Node
}

type LRUCache struct {
	cap         int
	cache       map[int]*Node
	left, right *Node
}

func Constructor(capacity int) LRUCache {
	left, right := &Node{Key: 0, Val: 0}, &Node{Key: 0, Val: 0}
	left.Next = right
	right.Prev = left
	return LRUCache{
		cap:   capacity,
		cache: make(map[int]*Node),
		left:  left,
		right: right,
	}
}

func (lru *LRUCache) remove(node *Node) {
	prev, next := node.Prev, node.Next
	prev.Next, next.Prev = next, prev
}

func (lru *LRUCache) add(node *Node) {
	prev, next := lru.right.Prev, lru.right
	prev.Next = node
	next.Prev = node
	node.Next, node.Prev = next, prev
}

func (lru *LRUCache) Get(key int) int {
	if v, ok := lru.cache[key]; ok {
		lru.remove(v)
		lru.add(v)
		return v.Val
	}
	return -1
}

func (lru *LRUCache) Put(key int, value int) {
	if v, ok := lru.cache[key]; ok {
		lru.remove(v)
	}

	lru.cache[key] = &Node{Key: key, Val: value}
	lru.add(lru.cache[key])

	if len(lru.cache) > lru.cap {
		nextC := lru.left.Next
		lru.remove(nextC)
		delete(lru.cache, nextC.Key)
	}
}

/**
 * Your LRUCache object will be instantiated and called as such:
 * obj := Constructor(capacity);
 * param_1 := obj.Get(key);
 * obj.Put(key,value);
 */
