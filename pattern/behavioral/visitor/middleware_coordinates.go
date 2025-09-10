package main

import "fmt"

// MiddleCoordinates is a concrete visitor
type MiddleCoordinates struct {
	x int
	y int
}

// visitForSquare is the method that calculates middle point coordinates for square
func (a *MiddleCoordinates) visitForSquare(s *Square) {
	// Calculate middle point coordinates for square.
	// Then assign in to the x and y instance variable.
	fmt.Println("Calculating middle point coordinates for square")
}

// visitForCircle is the method that calculates middle point coordinates for circle
func (a *MiddleCoordinates) visitForCircle(c *Circle) {
	fmt.Println("Calculating middle point coordinates for circle")
}

// visitForrectangle is the method that calculates middle point coordinates for rectangle
func (a *MiddleCoordinates) visitForrectangle(t *Rectangle) {
	fmt.Println("Calculating middle point coordinates for rectangle")
}
