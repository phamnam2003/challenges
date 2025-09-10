package main

// musket is a concrete implementation of the Gun interface
type musket struct {
	Gun
}

// newMusket is a factory method that creates and returns a new musket instance
func newMusket() IGun {
	return &musket{
		Gun: Gun{
			name:  "Musket gun",
			power: 1,
		},
	}
}
