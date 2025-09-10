package main

import "fmt"

// PassengerTrain is a concrete colleague
type PassengerTrain struct {
	mediator Mediator
}

// arrive tries to arrive
func (g *PassengerTrain) arrive() {
	if !g.mediator.canArrive(g) {
		fmt.Println("PassengerTrain: Arrival blocked, waiting")
		return
	}
	fmt.Println("PassengerTrain: Arrived")
}

// depart leaves the station
func (g *PassengerTrain) depart() {
	fmt.Println("PassengerTrain: Leaving")
	g.mediator.notifyAboutDeparture()
}

// permitArrival is called by the mediator to allow arrival
func (g *PassengerTrain) permitArrival() {
	fmt.Println("PassengerTrain: Arrival permitted, arriving")
	g.arrive()
}
