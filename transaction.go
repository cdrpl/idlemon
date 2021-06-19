package main

import "log"

// Represents a resource transaction.
type Transaction struct {
	Type   int `json:"type"`
	Amount int `json:"amount"`
}

// Will apply the transaction to a user.
func (r Transaction) Apply(user *User) {
	switch r.Type {
	case TRANSACTION_GEMS:
		user.Data.Resources[RESOURCE_GEMS].Amount += r.Amount

	default:
		log.Fatalf("failed to apply transaction of type %v, not handled in switch statement\n", r.Type)
	}
}
