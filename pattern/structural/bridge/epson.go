package main

import "fmt"

// Epson is a concrete implementation of the Printer interface
type Epson struct{}

// PrintFile prints a file using an Epson printer
func (p *Epson) PrintFile() {
	fmt.Println("Printing by a EPSON Printer")
}
