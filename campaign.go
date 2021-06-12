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

// Returns the amount of resources that can be collected and sets last collected at to now.
func (c *Campaign) Collect() (exp int, gold int, expStone int) {
	now := time.Now()
	timeDiff := int(now.Sub(c.LastCollectedAt).Seconds())

	// limit stockpile to 24 hours
	max := int((time.Hour * 24).Seconds())
	if timeDiff > max {
		timeDiff = max
	}

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

func GetCampaignLock(ctx context.Context, db *sql.Tx, userID int) (Campaign, error) {
	campaign := Campaign{UserID: userID}

	query := "SELECT id, level, last_collected_at FROM campaign WHERE user_id = ? FOR UPDATE"
	err := db.QueryRowContext(ctx, query, userID).Scan(&campaign.ID, &campaign.Level, &campaign.LastCollectedAt)

	return campaign, err
}
