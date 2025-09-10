package main

// Shape is the interface that defines the methods for a shape
type Shape interface {
	getType() string
	accept(Visitor)
}
