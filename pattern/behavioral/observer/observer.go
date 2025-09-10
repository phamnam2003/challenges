package main

// Observer is the observer interface
type Observer interface {
	update(string)
	getID() string
}
