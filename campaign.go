package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Campaign struct {
	Id              int       `json:"id"`
	UserId          int       `json:"-"`
	Level           int       `json:"level"`
	LastCollectedAt time.Time `json:"lastCollectedAt"`
}

// Will set campaign last collected at in the database.
func (c *Campaign) UpdateLastCollectedAt(ctx context.Context, tx pgx.Tx) error {
	query := "UPDATE campaign SET last_collected_at = $1 WHERE id = $2"

	_, err := tx.Exec(ctx, query, c.LastCollectedAt, c.Id)
	if err != nil {
		return fmt.Errorf("fail to update campaign row: %w", err)
	}

	return nil
}

// Will update the database to reflect collection of the campaign resources.
func (c *Campaign) Collect(ctx context.Context, tx pgx.Tx) ([3]Transaction, error) {
	timeDiff := time.Since(c.LastCollectedAt)
	timeDiffSec := int(timeDiff.Seconds())

	var transactions [3]Transaction
	transactions[0] = Transaction{Type: TRANSACTION_USER_EXP}
	transactions[1] = Transaction{Type: TRANSACTION_GOLD}
	transactions[2] = Transaction{Type: TRANSACTION_EXP_STONES}

	log.Println(c.Level)

	// min time interval between collections
	if timeDiff < time.Second {
		return transactions, nil
	}

	// limit stockpile to 24 hours
	if timeDiff > CAMPAIGN_MAX_COLLECT {
		timeDiff = CAMPAIGN_MAX_COLLECT
	}

	// calculate rewards
	exp := timeDiffSec * (CAMPAIGN_EXP_PER_SEC + (c.Level / 5 * CAMPAIGN_EXP_GROWTH))
	gold := timeDiffSec * (CAMPAIGN_GOLD_PER_SEC + (c.Level / 5 * CAMPAIGN_GOLD_GROWTH))
	expStones := timeDiffSec * (CAMPAIGN_EXP_STONE_PER_SEC + (c.Level / 5 * CAMPAIGN_EXP_STONE_GROWTH))

	c.LastCollectedAt = time.Now().UTC().Round(time.Second)

	if err := c.UpdateLastCollectedAt(ctx, tx); err != nil {
		return transactions, fmt.Errorf("fail to update campaign last collected at: %w", err)
	}

	if err := IncUserExp(ctx, tx, c.UserId, exp); err != nil {
		return transactions, fmt.Errorf("fail to increase user exp: %w", err)
	}

	if err := IncResource(ctx, tx, c.UserId, RESOURCE_GOLD, gold); err != nil {
		return transactions, fmt.Errorf("fail to increase gold resource: %w", err)
	}

	if err := IncResource(ctx, tx, c.UserId, RESOURCE_EXP_STONE, expStones); err != nil {
		return transactions, fmt.Errorf("fail to increase exp stone resource: %w", err)
	}

	transactions[0].Amount = exp
	transactions[1].Amount = gold
	transactions[2].Amount = expStones

	return transactions, nil
}

func FindCampaign(ctx context.Context, db *pgxpool.Pool, userId int) (Campaign, error) {
	var campaign Campaign

	query := "SELECT id, level, last_collected_at FROM campaign WHERE user_id = $1"
	err := db.QueryRow(ctx, query, userId).Scan(&campaign.Id, &campaign.Level, &campaign.LastCollectedAt)
	if err != nil {
		return campaign, fmt.Errorf("fail to query campaign row: %w", err)
	}

	return campaign, nil
}

func InsertCampaign(ctx context.Context, tx pgx.Tx, userId int) error {
	now := time.Now().UTC().Round(time.Second)

	query := "INSERT INTO campaign (user_id, last_collected_at) VALUES ($1, $2)"
	_, err := tx.Exec(ctx, query, userId, now)
	if err != nil {
		return err
	}

	return nil
}
