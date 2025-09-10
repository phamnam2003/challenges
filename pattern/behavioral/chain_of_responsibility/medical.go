package main

import "fmt"

// Medical struct department
type Medical struct {
	next Department
}

// execute medical department task
func (m *Medical) execute(p *Patient) {
	if p.medicineDone {
		fmt.Println("Medicine already given to patient")
		m.next.execute(p)
		return
	}
	fmt.Println("Medical giving medicine to patient")
	p.medicineDone = true
	m.next.execute(p)
}

// setNext department
func (m *Medical) setNext(next Department) {
	m.next = next
}
