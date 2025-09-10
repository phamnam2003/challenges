package main

// Mediator is the interface that defines the mediator's behavior
type Mediator interface {
	canArrive(Train) bool
	notifyAboutDeparture()
}
