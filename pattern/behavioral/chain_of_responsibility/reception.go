package main

import "fmt"

// Reception struct
type Reception struct {
	next Department
}

// execute reception department task
func (r *Reception) execute(p *Patient) {
	if p.registrationDone {
		fmt.Println("Patient registration already done")
		r.next.execute(p)
		return
	}
	fmt.Println("Reception registering patient")
	p.registrationDone = true
	r.next.execute(p)
}

// setNext department
func (r *Reception) setNext(next Department) {
	r.next = next
}
