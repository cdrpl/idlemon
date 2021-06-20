package main

import (
	"context"

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

	resources[RESOURCE_GOLD] = Resource{Id: RESOURCE_GOLD}
	resources[RESOURCE_GEMS] = Resource{Id: RESOURCE_GEMS}
	resources[RESOURCE_EXP_STONE] = Resource{Id: RESOURCE_EXP_STONE}
	resources[RESOURCE_EVO_STONE] = Resource{Id: RESOURCE_EVO_STONE}

	return resources
}

// Find all resources belonging to a user.
func FindResources(ctx context.Context, db *pgxpool.Pool, userId int) ([]Resource, error) {
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
