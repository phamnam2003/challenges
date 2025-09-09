// Package linkedlist provides a generic linked list implementation.
package linkedlist

// LinkedList generic linked list type
type LinkedList[T any] struct {
	V    T
	Next *LinkedList[T]
}
