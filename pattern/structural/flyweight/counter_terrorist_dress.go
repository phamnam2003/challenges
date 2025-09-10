package main

// CounterTerroristDress contains properties
type CounterTerroristDress struct {
	color string
}

// getColor return color of dress
func (c *CounterTerroristDress) getColor() string {
	return c.color
}

// newCounterTerroristDress is a constructor
func newCounterTerroristDress() *CounterTerroristDress {
	return &CounterTerroristDress{color: "green"}
}
