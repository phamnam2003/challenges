package main

import "fmt"

// Windows is a concrete device
type Windows struct{}

// insertIntoUSBPort simulates inserting a USB connector into a Windows machine
func (w *Windows) insertIntoUSBPort() {
	fmt.Println("USB connector is plugged into windows machine.")
}
