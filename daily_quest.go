package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
)

// Daily quest IDs.
const (
	DAILY_QUEST_SIGN_IN = iota + 1
)

type DailyQuest struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Required    int    `json:"required"`
}

func InsertDailyQuests(ctx context.Context, db *sql.DB, dc *DataCache) {
	for _, dailyQuest := range dc.DailyQuests {
		var id int

		err := db.QueryRowContext(ctx, "SELECT id FROM daily_quest WHERE id = ? FOR UPDATE", dailyQuest.ID).Scan(&id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				query := "INSERT INTO daily_quest (id, description, required) VALUES (?, ?, ?)"
				_, err := db.ExecContext(ctx, query, dailyQuest.ID, dailyQuest.Description, dailyQuest.Required)
				if err != nil {
					log.Fatalln("insert daily quests error:", err)
				}

				log.Printf("insert daily quest %+v\n", dailyQuest)
			} else {
				log.Fatalln("insert daily quests error:", err)
			}
		}
	}
}

func UnmarshallDailyQuestsJson() ([]DailyQuest, error) {
	var data map[string][]DailyQuest

	err := json.Unmarshal([]byte(dailyQuestsJson), &data)
	if err != nil {
		return nil, err
	}

	return data["dailyQuests"], nil
}
