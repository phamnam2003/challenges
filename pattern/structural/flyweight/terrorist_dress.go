package main

// TerroristDress contains properties of dress
type TerroristDress struct {
	color string
}

// getColor return color of TerroristDress
func (t *TerroristDress) getColor() string {
	return t.color
}

// newTerroristDress is a constructor for TerroristDress
func newTerroristDress() *TerroristDress {
	return &TerroristDress{color: "red"}
}
