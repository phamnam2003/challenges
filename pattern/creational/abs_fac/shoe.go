package main

// IShoe is the interface that defines the methods for a shoe
type IShoe interface {
	setLogo(logo string)
	setSize(size int)
	getLogo() string
	getSize() int
}

// Shoe is the struct that implements the IShoe interface
type Shoe struct {
	logo string
	size int
}

// setLogo sets the logo of the shoe
func (s *Shoe) setLogo(logo string) {
	s.logo = logo
}

// getLogo gets the logo of the shoe
func (s *Shoe) getLogo() string {
	return s.logo
}

// setSize sets the size of the shoe
func (s *Shoe) setSize(size int) {
	s.size = size
}

// getSize gets the size of the shoe
func (s *Shoe) getSize() int {
	return s.size
}
