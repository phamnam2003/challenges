package main

// IGun is the interface that defines the methods for a gun
type IGun interface {
	setName(name string)
	setPower(power int)
	getName() string
	getPower() int
}

// Gun is the struct that implements the IGun interface
type Gun struct {
	name  string
	power int
}

// setName sets the name of the gun
func (g *Gun) setName(name string) {
	g.name = name
}

// getName gets the name of the gun
func (g *Gun) getName() string {
	return g.name
}

// setPower sets the power of the gun
func (g *Gun) setPower(power int) {
	g.power = power
}

// getPower gets the power of the gun
func (g *Gun) getPower() int {
	return g.power
}
