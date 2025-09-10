package main

import "fmt"

// Item is the subject being observed
type Item struct {
	observerList []Observer
	name         string
	inStock      bool
}

// newItem creates a new item
func newItem(name string) *Item {
	return &Item{
		name: name,
	}
}

// updateAvailability updates the availability of the item and notifies all observers
func (i *Item) updateAvailability() {
	fmt.Printf("Item %s is now in stock\n", i.name)
	i.inStock = true
	i.notifyAll()
}

// register adds an observer to the observer list
func (i *Item) register(o Observer) {
	i.observerList = append(i.observerList, o)
}

// deregister removes an observer from the observer list
func (i *Item) deregister(o Observer) {
	i.observerList = removeFromslice(i.observerList, o)
}

// notifyAll notifies all observers about the item's availability
func (i *Item) notifyAll() {
	for _, observer := range i.observerList {
		observer.update(i.name)
	}
}

// removeFromslice removes an observer from the observer list
func removeFromslice(observerList []Observer, observerToRemove Observer) []Observer {
	observerListLength := len(observerList)
	for i, observer := range observerList {
		if observerToRemove.getID() == observer.getID() {
			observerList[observerListLength-1], observerList[i] = observerList[i], observerList[observerListLength-1]
			return observerList[:observerListLength-1]
		}
	}
	return observerList
}
