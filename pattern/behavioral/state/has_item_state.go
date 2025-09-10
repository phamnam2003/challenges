package main

import "fmt"

// HasItemState is a concrete state
type HasItemState struct {
	vendingMachine *VendingMachine
}

// requestItem requests an item
func (i *HasItemState) requestItem() error {
	if i.vendingMachine.itemCount == 0 {
		i.vendingMachine.setState(i.vendingMachine.noItem)
		return fmt.Errorf("no item present")
	}
	fmt.Printf("Item requestd\n")
	i.vendingMachine.setState(i.vendingMachine.itemRequested)
	return nil
}

// addItem adds items to the vending machine
func (i *HasItemState) addItem(count int) error {
	fmt.Printf("%d items added\n", count)
	i.vendingMachine.incrementItemCount(count)
	return nil
}

// insertMoney inserts money into the vending machine
func (i *HasItemState) insertMoney(money int) error {
	return fmt.Errorf("please select item first")
}

// dispenseItem dispenses an item
func (i *HasItemState) dispenseItem() error {
	return fmt.Errorf("please select item first")
}
