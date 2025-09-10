package main

import "fmt"

// Mac is a concrete device
type Mac struct{}

// InsertIntoLightningPort plugs the lightning connector into the mac machine
func (m *Mac) InsertIntoLightningPort() {
	fmt.Println("Lightning connector is plugged into mac machine.")
}
