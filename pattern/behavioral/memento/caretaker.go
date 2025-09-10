package main

// Caretaker is responsible for keeping the mementos
type Caretaker struct {
	mementoArray []*Memento
}

// addMemento adds a memento to the caretaker's list
func (c *Caretaker) addMemento(m *Memento) {
	c.mementoArray = append(c.mementoArray, m)
}

// getMemento retrieves a memento by index
func (c *Caretaker) getMemento(index int) *Memento {
	return c.mementoArray[index]
}
