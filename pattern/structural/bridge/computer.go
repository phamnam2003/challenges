package main

// Computer is the abstraction interface
type Computer interface {
	Print()
	SetPrinter(Printer)
}
