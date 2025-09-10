package main

import "fmt"

// Customer is a concrete observer
type Customer struct {
	id string
}

// update sends an email to the customer
func (c *Customer) update(itemName string) {
	fmt.Printf("Sending email to customer %s for item %s\n", c.id, itemName)
}

// getID returns the customer id
func (c *Customer) getID() string {
	return c.id
}
