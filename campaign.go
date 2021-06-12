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
	timeDiff := time.Since(c.LastCollectedAt)
	timeDiffSec := int(timeDiff.Seconds())

	// limit stockpile to 24 hours
	if timeDiff > CAMPAIGN_MAX_COLLECT {
		timeDiff = CAMPAIGN_MAX_COLLECT
	}

	if timeDiff > time.Second {
		exp = timeDiffSec * (CAMPAIGN_EXP_PER_SEC + (c.Level / 5 * CAMPAIGN_EXP_GROWTH))
		gold = timeDiffSec * (CAMPAIGN_GOLD_PER_SEC + (c.Level / 5 * CAMPAIGN_GOLD_GROWTH))
		expStone = timeDiffSec * (CAMPAIGN_EXP_STONE_PER_SEC + (c.Level / 5 * CAMPAIGN_EXP_STONE_GROWTH))
		c.LastCollectedAt = time.Now()
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
