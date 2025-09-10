package main

import "fmt"

// Windows is a concrete implementer
type Windows struct {
	printer Printer
}

// Print calls the PrintFile method of the printer
func (w *Windows) Print() {
	fmt.Println("Print request for windows")
	w.printer.PrintFile()
}

// SetPrinter sets the printer for the windows
func (w *Windows) SetPrinter(p Printer) {
	w.printer = p
}
