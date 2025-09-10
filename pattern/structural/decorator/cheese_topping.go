package main

// CheeseTopping is a concrete decorator that adds cheese topping to the pizza
type CheeseTopping struct {
	pizza IPizza
}

// getPrice returns the price of the pizza with cheese topping
func (c *CheeseTopping) getPrice() int {
	pizzaPrice := c.pizza.getPrice()
	return pizzaPrice + 10
}
