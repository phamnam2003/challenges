package main

// Originator is the object that holds the state
type Originator struct {
	state string
}

// createMemento creates a memento with the current state
func (e *Originator) createMemento() *Memento {
	return &Memento{state: e.state}
}

// restoreMemento restores the state from the memento
func (e *Originator) restoreMemento(m *Memento) {
	e.state = m.getSavedState()
}

// setState sets the state of the originator
func (e *Originator) setState(state string) {
	e.state = state
}

// getState gets the current state of the originator
func (e *Originator) getState() string {
	return e.state
}
