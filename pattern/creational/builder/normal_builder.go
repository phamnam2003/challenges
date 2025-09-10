package main

// NormalBuilder is a concrete builder that implements the Builder interface
type NormalBuilder struct {
	windowType string
	doorType   string
	floor      int
}

// newNormalBuilder creates a new instance of NormalBuilder
func newNormalBuilder() *NormalBuilder {
	return &NormalBuilder{}
}

// setWindowType sets the window type for the house
func (b *NormalBuilder) setWindowType() {
	b.windowType = "Wooden Window"
}

// setDoorType sets the door type for the house
func (b *NormalBuilder) setDoorType() {
	b.doorType = "Wooden Door"
}

// setNumFloor sets the number of floors for the house
func (b *NormalBuilder) setNumFloor() {
	b.floor = 2
}

// getHouse returns the constructed house
func (b *NormalBuilder) getHouse() House {
	return House{
		doorType:   b.doorType,
		windowType: b.windowType,
		floor:      b.floor,
	}
}
