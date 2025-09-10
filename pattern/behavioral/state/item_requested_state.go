package main

import "fmt"

// ItemRequestedState is the state when item is requested
type ItemRequestedState struct {
	vendingMachine *VendingMachine
}

// requestItem requests an item
func (i *ItemRequestedState) requestItem() error {
	return fmt.Errorf("item already requested")
}

// addItem adds item to the vending machine
func (i *ItemRequestedState) addItem(count int) error {
	return fmt.Errorf("item Dispense in progress")
}

// insertMoney inserts money into the vending machine
func (i *ItemRequestedState) insertMoney(money int) error {
	if money < i.vendingMachine.itemPrice {
		return fmt.Errorf("inserted money is less. Please insert %d", i.vendingMachine.itemPrice)
	}
	fmt.Println("Money entered is ok")
	i.vendingMachine.setState(i.vendingMachine.hasMoney)
	return nil
}

// dispenseItem dispenses the item
func (i *ItemRequestedState) dispenseItem() error {
	return fmt.Errorf("please insert money first")
}
