package main

// Player represents a player in the game
type Player struct {
	dress      Dress
	playerType string
	lat        int
	long       int
}

// newPlayer creates a new player with the given type and dress type
func newPlayer(playerType, dressType string) *Player {
	dress, _ := getDressFactorySingleInstance().getDressByType(dressType)
	return &Player{
		playerType: playerType,
		dress:      dress,
	}
}

// newLocation sets a new location for the player
func (p *Player) newLocation(lat, long int) {
	p.lat = lat
	p.long = long
}
