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
	Count           int       `json:"count"`
	LastCompletedAt time.Time `json:"lastCompletedAt"`
}

// Will check if the quest has already been completed today.
func (udq UserDailyQuest) IsCompleted() bool {
	y, m, d := time.Now().UTC().Date()
	startOfToday := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)

	return udq.LastCompletedAt.Unix() >= startOfToday.Unix()
}

func CreateUserDailyQuest() UserDailyQuest {
	// set before the start of today so the user can complete the quest even if they signed up today
	lastCompletedAt := time.Now().Add(-time.Hour * 48).UTC().Round(time.Second)

	return UserDailyQuest{LastCompletedAt: lastCompletedAt}
}

func CompleteDailyQuest(id int, user *User) Reward {
	user.Data.DailyQuests[id].Count = 0
	user.Data.DailyQuests[id].LastCompletedAt = time.Now().UTC().Round(time.Second)

	switch id {
	case DAILY_QUEST_SIGN_IN:
		user.Data.Resources[RESOURCE_GEMS].Amount += DAILY_SIGN_IN_REWARD
		return Reward{Type: GEMS, Amount: DAILY_SIGN_IN_REWARD}

	default:
		log.Printf("attempt to complete daily quest not handled by switch statement: %v\n", id)
	}

	return Reward{}
}
