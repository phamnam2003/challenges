package main

import "errors"

// ISportsFactory defines the interface for a sports factory
type ISportsFactory interface {
	makeShoe() IShoe
	makeShirt() IShirt
}

// GetSportsFactory returns a sports factory based on the brand
func GetSportsFactory(brand string) (ISportsFactory, error) {
	switch brand {
	case "nike":
		{
			return &Nike{}, nil
		}
	case "adidas":
		{
			return &Adidas{}, nil
		}
	default:
		{
			return nil, errors.New("not mapping type")
		}
	}
}
