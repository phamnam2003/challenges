package main

import "fmt"

// Hp is a concrete implementation of the Printer interface
type Hp struct{}

// PrintFile prints a file using a HP printer
func (p *Hp) PrintFile() {
	fmt.Println("Printing by a HP Printer")
}
