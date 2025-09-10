package main

// IShirt is the interface that defines the methods for a shirt
type IShirt interface {
	setLogo(logo string)
	setSize(size int)
	getLogo() string
	getSize() int
}

// Shirt is the concrete implementation of IShirt
type Shirt struct {
	logo string
	size int
}

// setLogo sets the logo of the shirt
func (s *Shirt) setLogo(logo string) {
	s.logo = logo
}

// getLogo gets the logo of the shirt
func (s *Shirt) getLogo() string {
	return s.logo
}

// setSize sets the size of the shirt
func (s *Shirt) setSize(size int) {
	s.size = size
}

// getSize gets the size of the shirt
func (s *Shirt) getSize() int {
	return s.size
}
