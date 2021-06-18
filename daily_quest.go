package main

import (
	"encoding/json"
	"log"
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
	Count       int  `json:"count"`
	IsCompleted bool `json:"isCompleted"`
}

func CompleteDailyQuest(id int, user *User) Reward {
	user.Data.DailyQuests[id].Count = 0
	user.Data.DailyQuests[id].IsCompleted = true

	switch id {
	case DAILY_QUEST_SIGN_IN:
		user.Data.Resources[RESOURCE_GEMS].Amount += DAILY_SIGN_IN_REWARD
		return Reward{Type: GEMS, Amount: DAILY_SIGN_IN_REWARD}

	default:
		log.Printf("attempt to complete daily quest not handled by switch statement: %v\n", id)
	}

	return Reward{}
}
