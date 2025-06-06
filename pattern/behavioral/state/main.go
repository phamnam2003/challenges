package main

import (
	"fmt"
	"log"
)

func main() {
	vendingMachine := newVendingMachine(1, 10)

	err := vendingMachine.requestItem()
	if err != nil {
		log.Fatal(err)
	}

	err = vendingMachine.insertMoney(10)
	if err != nil {
		log.Fatal(err)
	}

	err = vendingMachine.dispenseItem()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println()

	err = vendingMachine.addItem(2)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println()

	err = vendingMachine.requestItem()
	if err != nil {
		log.Fatal()
	}

	err = vendingMachine.insertMoney(10)
	if err != nil {
		log.Fatal(err)
	}

	err = vendingMachine.dispenseItem()
	if err != nil {
		log.Fatal(err)
	}
}
