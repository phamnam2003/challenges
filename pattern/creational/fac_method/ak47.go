package main

// Ak47 is a concrete implementation of IGun
type Ak47 struct {
	Gun
}

// newAk47 is a constructor for Ak47
func newAk47() IGun {
	return &Ak47{
		Gun: Gun{
			name:  "AK47 gun",
			power: 4,
		},
	}
}
