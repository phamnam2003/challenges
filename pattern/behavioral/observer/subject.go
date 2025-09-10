package main

// Subject is the interface that defines the methods for a subject
type Subject interface {
	register(observer Observer)
	deregister(observer Observer)
	notifyAll()
}
