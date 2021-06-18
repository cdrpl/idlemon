package main

import (
	"encoding/json"
	"log"
	"time"
)

type DailyQuest struct {
	Description string `json:"description"`
	Required    int    `json:"required"`
}

func UnmarshallDailyQuestsJson() ([]DailyQuest, error) {
	var data map[string][]DailyQuest

	err := json.Unmarshal([]byte(dailyQuestsJson), &data)
	if err != nil {
		return nil, err
	}

	return data["dailyQuests"], nil
}

type UserDailyQuest struct {
	ID              int       `json:"id"`
	Count           int       `json:"count"`
	LastCompletedAt time.Time `json:"lastCompletedAt"`
}

func CreateUserDailyQuest(id int) UserDailyQuest {
	// set before the start of today so the user can complete the quest even if they signed up today
	lastCompletedAt := time.Now().Add(-time.Hour * 48).UTC().Round(time.Second)

	return UserDailyQuest{ID: id, LastCompletedAt: lastCompletedAt}
}

// Will check if the quest has already been completed today.
func (udq *UserDailyQuest) IsCompleted() bool {
	y, m, d := time.Now().UTC().Date()
	startOfToday := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)

	return udq.LastCompletedAt.Unix() >= startOfToday.Unix()
}

// Will increase the count if the quest hasn't been completed today.
func (udq *UserDailyQuest) IncreaseCount() {
	if !udq.IsCompleted() {
		udq.Count++
	}
}

// Will complete the daily quest if it hasn't been completed and the requirements are met.
func (udq *UserDailyQuest) Complete() Reward {
	udq.Count = 0
	udq.LastCompletedAt = time.Now().UTC().Round(time.Second)

	switch udq.ID {
	case DAILY_QUEST_SIGN_IN:
		return Reward{Type: GEMS, Amount: DAILY_SIGN_IN_REWARD}

	default:
		log.Printf("attempt to complete daily quest not handled by switch statement: %v\n", udq.ID)
	}

	return Reward{}
}
