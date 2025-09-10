package main

// VeggieMania is a concrete component
type VeggieMania struct{}

// getPrice returns the price of VeggieMania
func (p *VeggieMania) getPrice() int {
	return 15
}
