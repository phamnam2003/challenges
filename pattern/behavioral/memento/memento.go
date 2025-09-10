package main

// Memento is the memento struct that stores the state of the Originator
type Memento struct {
	state string
}

// getSavedState returns the saved state
func (m *Memento) getSavedState() string {
	return m.state
}
