package main

import "fmt"

// Account is a struct that represents a account in system
type Account struct {
	name string
}

// newAccount is a constructor for Account
func newAccount(accountName string) *Account {
	return &Account{
		name: accountName,
	}
}

// checkAccount checks if the account name is correct
func (a *Account) checkAccount(accountName string) error {
	if a.name != accountName {
		return fmt.Errorf("Account Name is incorrect")
	}
	fmt.Println("Account Verified")
	return nil
}
