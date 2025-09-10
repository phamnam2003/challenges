package main

// Square is a concrete element
type Square struct {
	side int
}

// accept accepts a visitor
func (s *Square) accept(v Visitor) {
	v.visitForSquare(s)
}

// getType returns the type of the shape
func (s *Square) getType() string {
	return "Square"
}
