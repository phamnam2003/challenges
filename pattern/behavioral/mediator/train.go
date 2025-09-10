package main

// Train is the interface that defines the methods for a train
type Train interface {
	arrive()
	depart()
	permitArrival()
}
