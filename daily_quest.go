package main

import (
	"encoding/json"
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
