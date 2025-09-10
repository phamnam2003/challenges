package main

// game struct represents a game with counter-ter
type game struct {
	terrorists        []*Player
	counterTerrorists []*Player
}

// newGame creates a new game instance
func newGame() *game {
	return &game{
		terrorists:        make([]*Player, 1),
		counterTerrorists: make([]*Player, 1),
	}
}

// addTerror append terrorist to the game
func (c *game) addTerrorist(dressType string) {
	player := newPlayer("T", dressType)
	c.terrorists = append(c.terrorists, player)
}

// addCounterTerrorist append player with CT counterTerrorists
func (c *game) addCounterTerrorist(dressType string) {
	player := newPlayer("CT", dressType)
	c.counterTerrorists = append(c.counterTerrorists, player)
}
