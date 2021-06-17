package main

import (
	"context"
	"database/sql"
	"log"
)

type UserDailyQuest struct {
	ID           int `json:"id"`
	UserID       int `json:"userId"`
	DailyQuestID int `json:"dailyQuestId"`
	Count        int `json:"count"`
	IsCompleted  int `json:"isCompleted"`
}

// Update the daily quest for user sign in.
func SignInDailyQuest(ctx context.Context, db *sql.DB, userID int) error {
	query := "UPDATE user_daily_quest SET count = count + 1 WHERE (user_id = ? AND daily_quest_id = ? AND count = 0)"
	_, err := db.ExecContext(ctx, query, userID, DAILY_QUEST_SIGN_IN)

	return err
}

func CompleteDailyQuest(id int) Reward {
	switch id {
	case DAILY_QUEST_SIGN_IN:
		return Reward{Type: GEMS, Amount: DAILY_SIGN_IN_REWARD}
		break

	default:
		log.Printf("attempt to complete daily quest not handled by switch statement: %v\n", id)
	}
}

func CompleteSignInDailyQuest() {

}
