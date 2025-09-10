package main

// IglooBuilder is the concrete builder for constructing an igloo house
type IglooBuilder struct {
	windowType string
	doorType   string
	floor      int
}

// newIglooBuilder initializes a new IglooBuilder
func newIglooBuilder() *IglooBuilder {
	return &IglooBuilder{}
}

// setWindowType sets the window type for the igloo
func (b *IglooBuilder) setWindowType() {
	b.windowType = "Snow Window"
}

// setDoorType sets the door type for the igloo
func (b *IglooBuilder) setDoorType() {
	b.doorType = "Snow Door"
}

// setNumFloor sets the number of floors for the igloo
func (b *IglooBuilder) setNumFloor() {
	b.floor = 1
}

// getHouse constructs and returns the igloo house
func (b *IglooBuilder) getHouse() House {
	return House{
		doorType:   b.doorType,
		windowType: b.windowType,
		floor:      b.floor,
	}
}
