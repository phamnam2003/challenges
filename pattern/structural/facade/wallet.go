package main

import "fmt"

// Wallet is a struct that represents a wallet with a balance
type Wallet struct {
	balance int
}

// newWallet is a constructor function that returns a new instance of Wallet
func newWallet() *Wallet {
	return &Wallet{
		balance: 0,
	}
}

// creditBalance adds the amount to the wallet balance
func (w *Wallet) creditBalance(amount int) {
	w.balance += amount
	fmt.Println("Wallet balance added successfully")
}

// debitBalance deducts the amount from the wallet balance if sufficient balance is available
func (w *Wallet) debitBalance(amount int) error {
	if w.balance < amount {
		return fmt.Errorf("balance is not sufficient")
	}
	fmt.Println("Wallet balance is Sufficient")
	w.balance = w.balance - amount
	return nil
}
