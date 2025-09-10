package main

import "fmt"

// Doctor struct
type Doctor struct {
	next Department
}

// execute doctor department task
func (d *Doctor) execute(p *Patient) {
	if p.doctorCheckUpDone {
		fmt.Println("Doctor checkup already done")
		d.next.execute(p)
		return
	}
	fmt.Println("Doctor checking patient")
	p.doctorCheckUpDone = true
	d.next.execute(p)
}

// setNext department
func (d *Doctor) setNext(next Department) {
	d.next = next
}
