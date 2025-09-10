package main

import "fmt"

// VendingMachine is the struct stateful maker
type VendingMachine struct {
	hasItem       State
	itemRequested State
	hasMoney      State
	noItem        State

	currentState State

	itemCount int
	itemPrice int
}

// newVendingMachine is the constructor for VendingMachine
func newVendingMachine(itemCount, itemPrice int) *VendingMachine {
	v := &VendingMachine{
		itemCount: itemCount,
		itemPrice: itemPrice,
	}
	hasItemState := &HasItemState{
		vendingMachine: v,
	}
	itemRequestedState := &ItemRequestedState{
		vendingMachine: v,
	}
	hasMoneyState := &HasMoneyState{
		vendingMachine: v,
	}
	noItemState := &NoItemState{
		vendingMachine: v,
	}

	v.setState(hasItemState)
	v.hasItem = hasItemState
	v.itemRequested = itemRequestedState
	v.hasMoney = hasMoneyState
	v.noItem = noItemState
	return v
}

// requestItem requests an item from the vending machine
func (v *VendingMachine) requestItem() error {
	return v.currentState.requestItem()
}

// addItem adds items to the vending machine
func (v *VendingMachine) addItem(count int) error {
	return v.currentState.addItem(count)
}

// insertMoney inserts money into the vending machine
func (v *VendingMachine) insertMoney(money int) error {
	return v.currentState.insertMoney(money)
}

// dispenseItem dispenses an item from the vending machine
func (v *VendingMachine) dispenseItem() error {
	return v.currentState.dispenseItem()
}

// setState sets the current state of the vending machine
func (v *VendingMachine) setState(s State) {
	v.currentState = s
}

// insertItemIntoSlot decreases the item by count
func (v *VendingMachine) incrementItemCount(count int) {
	fmt.Printf("Adding %d items\n", count)
	v.itemCount = v.itemCount + count
}
