package main

import (
	"context"
	"database/sql"
	"time"
)

type Campaign struct {
	ID              int       `json:"id"`
	UserID          int       `json:"userId"`
	Level           int       `json:"level"`
	LastCollectedAt time.Time `json:"lastCollectedAt"`
}

func GetCampaign(ctx context.Context, db *sql.DB, userID int) (Campaign, error) {
	campaign := Campaign{UserID: userID}

	query := "SELECT id, level, last_collected_at FROM campaign WHERE user_id = ?"
	err := db.QueryRowContext(ctx, query, userID).Scan(&campaign.ID, &campaign.Level, &campaign.LastCollectedAt)

	return campaign, err
}
