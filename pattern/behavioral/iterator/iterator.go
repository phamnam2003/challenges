package main

// Iterator is the interface for iterating over a collection
type Iterator interface {
	hasNext() bool
	getNext() *User
}
