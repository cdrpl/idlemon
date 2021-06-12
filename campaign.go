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

// Returns the amount of resources collectible and sets last collected at to now.
func (c *Campaign) Collect() (exp int, gold int, expStone int) {
	now := time.Now()
	timeDiff := int(now.Sub(c.LastCollectedAt).Seconds())

	if timeDiff > 1 {
		exp = timeDiff * (CAMPAIGN_EXP + (c.Level / 5 * CAMPAIGN_INCREASE))
		gold = timeDiff * (CAMPAIGN_GOLD + (c.Level / 5 * CAMPAIGN_INCREASE))
		expStone = timeDiff * (CAMPAIGN_EXP_STONE + (c.Level / 5 * CAMPAIGN_INCREASE))
		c.LastCollectedAt = now
	}

	return
}

func GetCampaign(ctx context.Context, db *sql.DB, userID int) (Campaign, error) {
	campaign := Campaign{UserID: userID}

	query := "SELECT id, level, last_collected_at FROM campaign WHERE user_id = ?"
	err := db.QueryRowContext(ctx, query, userID).Scan(&campaign.ID, &campaign.Level, &campaign.LastCollectedAt)

	return campaign, err
}
