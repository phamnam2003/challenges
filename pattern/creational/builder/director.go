package main

// Director is responsible for managing the correct sequence of building steps
type Director struct {
	builder IBuilder
}

// newDirector creates a new Director with the given builder
func newDirector(b IBuilder) *Director {
	return &Director{
		builder: b,
	}
}

// setBuilder sets the builder for the director
func (d *Director) setBuilder(b IBuilder) {
	d.builder = b
}

// buildHouse constructs the house using the builder
func (d *Director) buildHouse() House {
	d.builder.setDoorType()
	d.builder.setWindowType()
	d.builder.setNumFloor()
	return d.builder.getHouse()
}
