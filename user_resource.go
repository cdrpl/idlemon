package main

import (
	"context"
	"database/sql"
)

type UserResource struct {
	ID         int `json:"id"`
	UserID     int `json:"userId"`
	ResourceID int `json:"resourceId"`
	Amount     int `json:"amount"`
}

func GetUserResources(ctx context.Context, db *sql.DB, userID int) ([]UserResource, error) {
	userResources := make([]UserResource, 0)

	rows, err := db.QueryContext(ctx, "SELECT * FROM user_resource WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		userResource := UserResource{}

		err := rows.Scan(&userResource.ID, &userResource.UserID, &userResource.ResourceID, &userResource.Amount)
		if err != nil {
			return nil, err
		}

		userResources = append(userResources, userResource)
	}

	return userResources, nil
}
