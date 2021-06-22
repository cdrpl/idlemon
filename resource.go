package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Represents a monetary resource such as gold.
type Resource struct {
	Id     int `json:"id"`
	Type   int `json:"type"`
	Amount int `json:"amount"`
}

// Return a slice of all resources in the game.
func Resources() []Resource {
	resources := make([]Resource, 4)

	resources[RESOURCE_GOLD] = Resource{Type: RESOURCE_GOLD}
	resources[RESOURCE_GEMS] = Resource{Type: RESOURCE_GEMS}
	resources[RESOURCE_EXP_STONE] = Resource{Type: RESOURCE_EXP_STONE}
	resources[RESOURCE_EVO_STONE] = Resource{Type: RESOURCE_EVO_STONE}

	return resources
}

// Find all resources belonging to a user.
func FindResources(ctx context.Context, db *pgxpool.Pool, userId uuid.UUID) ([]Resource, error) {
	resources := make([]Resource, 0)

	query := "SELECT id, type, amount FROM resources WHERE user_id = $1"
	rows, err := db.Query(ctx, query, userId)
	if err != nil {
		return resources, err
	}

	for rows.Next() {
		var resource Resource

		err := rows.Scan(&resource.Id, &resource.Type, &resource.Amount)
		if err != nil {
			return resources, err
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// Will find a resource for update.
func FindResourceLock(ctx context.Context, tx pgx.Tx, userId uuid.UUID, resourceType int) (Resource, error) {
	resource := Resource{Type: resourceType}

	query := "SELECT id, amount FROM resources WHERE (user_id = $1 AND type = $2) FOR UPDATE"
	err := tx.QueryRow(ctx, query, userId, resourceType).Scan(&resource.Id, &resource.Amount)
	if err != nil {
		return resource, fmt.Errorf("fail to query resources table: %w", err)
	}

	return resource, nil
}

// Will insert a resource row for every resource type.
func InsertResources(ctx context.Context, tx pgx.Tx, dc *DataCache, userId uuid.UUID) error {
	for _, resource := range dc.Resources {
		query := "INSERT INTO resources (user_id, type) VALUES ($1, $2)"

		_, err := tx.Exec(ctx, query, userId, resource.Type)
		if err != nil {
			return fmt.Errorf("fail to insert resource row of type %v: %w", resource.Type, err)
		}
	}

	return nil
}

// Will increase a resource row in the database by the specific amount.
func IncResource(ctx context.Context, tx pgx.Tx, userId uuid.UUID, resourceType int, amount int) error {
	query := "UPDATE resources SET amount = amount + $1 WHERE (user_id = $2 AND type = $3)"
	_, err := tx.Exec(ctx, query, amount, userId, resourceType)

	return err
}
