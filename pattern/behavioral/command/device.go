package main

// Device is the receiver
type Device interface {
	on()
	off()
}
