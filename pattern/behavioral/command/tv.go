package main

import "fmt"

// Tv is a concrete device
type Tv struct {
	isRunning bool
}

// on turns the tv on
func (t *Tv) on() {
	t.isRunning = true
	fmt.Println("Turning tv on")
}

// off turns the tv off
func (t *Tv) off() {
	t.isRunning = false
	fmt.Println("Turning tv off")
}
