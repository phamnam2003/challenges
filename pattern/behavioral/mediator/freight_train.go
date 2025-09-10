package main

import "fmt"

// FreightTrain is a concrete colleague
type FreightTrain struct {
	mediator Mediator
}

// arrive tries to arrive
func (g *FreightTrain) arrive() {
	if !g.mediator.canArrive(g) {
		fmt.Println("FreightTrain: Arrival blocked, waiting")
		return
	}
	fmt.Println("FreightTrain: Arrived")
}

// depart leaves the station
func (g *FreightTrain) depart() {
	fmt.Println("FreightTrain: Leaving")
	g.mediator.notifyAboutDeparture()
}

// permitArrival is called by the mediator to allow arrival
func (g *FreightTrain) permitArrival() {
	fmt.Println("FreightTrain: Arrival permitted")
	g.arrive()
}
