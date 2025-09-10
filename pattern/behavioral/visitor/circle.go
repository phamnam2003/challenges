package main

// Circle is a concrete element that implements the Shape interface
type Circle struct {
	radius int
}

// accept allows a visitor to visit the Circle
func (c *Circle) accept(v Visitor) {
	v.visitForCircle(c)
}

// getType returns the type of the shape
func (c *Circle) getType() string {
	return "Circle"
}
