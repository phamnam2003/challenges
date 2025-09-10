package main

// TomatoTopping is a concrete decorator that adds tomato topping to the pizza
type TomatoTopping struct {
	pizza IPizza
}

// getPrice returns the price of the pizza with tomato topping
func (c *TomatoTopping) getPrice() int {
	pizzaPrice := c.pizza.getPrice()
	return pizzaPrice + 7
}
