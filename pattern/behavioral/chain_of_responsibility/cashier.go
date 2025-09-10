package main

import "fmt"

// Cashier department
type Cashier struct {
	next Department
}

// execute cashier department task
func (c *Cashier) execute(p *Patient) {
	if p.paymentDone {
		fmt.Println("Payment Done")
	}
	fmt.Println("Cashier getting money from patient patient")
}

// set next department
func (c *Cashier) setNext(next Department) {
	c.next = next
}
