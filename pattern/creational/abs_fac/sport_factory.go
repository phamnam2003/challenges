package main

import "errors"

type ISportsFactory interface {
	makeShoe() IShoe
	makeShirt() IShirt
}

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
