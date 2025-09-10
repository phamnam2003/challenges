package main

import "fmt"

// WindowsAdapter is a concrete adapter
type WindowsAdapter struct {
	windowMachine *Windows
}

// InsertIntoLightningPort converts Lightning signal to USB and inserts into USB port
func (w *WindowsAdapter) InsertIntoLightningPort() {
	fmt.Println("Adapter converts Lightning signal to USB.")
	w.windowMachine.insertIntoUSBPort()
}
