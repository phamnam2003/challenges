package main

import (
	"errors"
)

// getGun is the factory method that returns the gun based on the type passed
func getGun(gunType string) (IGun, error) {
	switch gunType {
	case "ak47":
		{
			return newAk47(), nil
		}
	case "musket":
		{
			return newMusket(), nil
		}
	default:
		return nil, errors.New("wrong gun type passed")
	}
}
