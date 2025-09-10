package main

// Collection is the interface for creating an iterator
type Collection interface {
	createIterator() Iterator
}
