package main

// Department interface
type Department interface {
	execute(*Patient)
	setNext(Department)
}
