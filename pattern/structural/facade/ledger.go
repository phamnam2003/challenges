package main

import "fmt"

// Ledger is a subsystem that handles ledger entries
type Ledger struct{}

// makeEntry makes a ledger entry
func (s *Ledger) makeEntry(accountID, txnType string, amount int) {
	fmt.Printf("Make ledger entry for accountId %s with txnType %s for amount %d\n", accountID, txnType, amount)
}
