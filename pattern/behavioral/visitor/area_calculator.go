package main

import (
	"fmt"
)

// AreaCalculator is a concrete visitor
type AreaCalculator struct {
	area int
}

// visitForSquare is the method that calculates area for square
func (a *AreaCalculator) visitForSquare(s *Square) {
	// Calculate area for square.
	// Then assign in to the area instance variable.
	fmt.Println("Calculating area for square")
}

// visitForCircle is the method that calculates area for circle
func (a *AreaCalculator) visitForCircle(s *Circle) {
	fmt.Println("Calculating area for circle")
}

// visitForrectangle is the method that calculates area for rectangle
func (a *AreaCalculator) visitForrectangle(s *Rectangle) {
	fmt.Println("Calculating area for rectangle")
}
