package main

import "fmt"

// HasMoneyState is a concrete state
type HasMoneyState struct {
	vendingMachine *VendingMachine
}

// requestItem dispenses an item if there is enough money
func (i *HasMoneyState) requestItem() error {
	return fmt.Errorf("item dispense in progress")
}

// addItem adds items to the vending machine
func (i *HasMoneyState) addItem(count int) error {
	return fmt.Errorf("item dispense in progress")
}

// insertMoney returns an error since money has already been inserted
func (i *HasMoneyState) insertMoney(money int) error {
	return fmt.Errorf("item out of stock")
}

// dispenseItem dispenses an item and updates the state of the vending machine
func (i *HasMoneyState) dispenseItem() error {
	fmt.Println("Dispensing Item")
	i.vendingMachine.itemCount = i.vendingMachine.itemCount - 1
	if i.vendingMachine.itemCount == 0 {
		i.vendingMachine.setState(i.vendingMachine.noItem)
	} else {
		i.vendingMachine.setState(i.vendingMachine.hasItem)
	}
	return nil
}
