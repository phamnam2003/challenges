package main

import "fmt"

// Notification is the struct that handles notifications
type Notification struct{}

// snedWalletCreditNotification sends a wallet credit notification
func (n *Notification) sendWalletCreditNotification() {
	fmt.Println("Sending wallet credit notification")
}

// sendWalletDebitNotification sends a wallet debit notification
func (n *Notification) sendWalletDebitNotification() {
	fmt.Println("Sending wallet debit notification")
}
