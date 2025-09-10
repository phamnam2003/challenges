package main

// Rectangle is a concrete element
type Rectangle struct {
	l int
	b int
}

// accept accepts a visitor
func (t *Rectangle) accept(v Visitor) {
	v.visitForrectangle(t)
}

// getType returns the type of the shape
func (t *Rectangle) getType() string {
	return "rectangle"
}
