package main

import "fmt"

// Mac is a concrete abstraction
type Mac struct {
	printer Printer
}

// Print calls the PrintFile method of the printer
func (m *Mac) Print() {
	fmt.Println("Print request for mac")
	m.printer.PrintFile()
}

// SetPrinter sets the printer for the mac
func (m *Mac) SetPrinter(p Printer) {
	m.printer = p
}
