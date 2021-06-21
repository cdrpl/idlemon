package main

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

// Represents a modification of a integer value such as gold or user exp.
type Transaction struct {
	Type   int `json:"type"`
	Amount int `json:"amount"`
}

// Will apply the transaction to the correct table row.
func (r Transaction) Apply(ctx context.Context, tx pgx.Tx, userId uuid.UUID) error {
	switch r.Type {
	case TRANSACTION_GEMS:
		return IncResource(ctx, tx, userId, r.Type, r.Amount)

	default:
		log.Fatalf("failed to apply transaction of type %v, not handled in switch statement\n", r.Type)
	}

	return nil
}
