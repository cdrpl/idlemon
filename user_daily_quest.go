package main

import (
	"context"
	"database/sql"
	"fmt"
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
	query := "UPDATE user_daily_quest SET count = count + 1 WHERE (daily_quest_id = ? AND user_id = ?)"
	result, err := db.ExecContext(ctx, query, DAILY_QUEST_SIGN_IN, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	} else if rowsAffected == 0 {
		return fmt.Errorf("update daily quest %v for user %v failed, rows affected was 0", DAILY_QUEST_SIGN_IN, userID)
	}

	return nil
}
