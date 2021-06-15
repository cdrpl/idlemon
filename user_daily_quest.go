package main

import (
	"context"
	"database/sql"
)

type UserDailyQuest struct {
	ID           int `json:"id"`
	UserID       int `json:"userId"`
	DailyQuestID int `json:"dailyQuestId"`
	Count        int `json:"count"`
	IsCollected  int `json:"isCollected"`
}

// Update the daily quest for user sign in.
func SignInDailyQuest(ctx context.Context, db *sql.DB, userID int) error {
	query := "UPDATE user_daily_quest SET count = count + 1 WHERE (daily_quest_id = ? AND user_id = ? AND count = 0)"
	_, err := db.ExecContext(ctx, query, DAILY_QUEST_SIGN_IN, userID)

	return err
}
