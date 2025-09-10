package main

import "fmt"

// NoItemState is a concrete state
type NoItemState struct {
	vendingMachine *VendingMachine
}

// requestItem handles the request item action
func (i *NoItemState) requestItem() error {
	return fmt.Errorf("item out of stock")
}

// addItem handles the add item action
func (i *NoItemState) addItem(count int) error {
	i.vendingMachine.incrementItemCount(count)
	i.vendingMachine.setState(i.vendingMachine.hasItem)
	return nil
}

// insertMoney handles the insert money action
func (i *NoItemState) insertMoney(money int) error {
	return fmt.Errorf("item out of stock")
}

// dispenseItem handles the dispense item action
func (i *NoItemState) dispenseItem() error {
	return fmt.Errorf("item out of stock")
}
