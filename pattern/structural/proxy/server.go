package main

// handler is the interface that defines the methods for handling requests
type server interface {
	handleRequest(string, string) (int, string)
}
