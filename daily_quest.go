package main

import (
	"time"
)

type DailyQuest struct {
	Required int    `json:"required"`
	Reward   Reward `json:"reward"`
}

// Returns a slice of all the daily quests
func DailyQuests() []DailyQuest {
	dailyQuests := make([]DailyQuest, 1)

	// require user to sign in
	dailyQuests[DAILY_QUEST_SIGN_IN] = DailyQuest{
		Required: 1,
		Reward: Reward{
			Type:   REWARD_GEMS,
			Amount: 1000,
		},
	}

	return dailyQuests
}

type DailyQuestProgress struct {
	DailyQuestID    int       `json:"dailyQuestId"`
	Count           int       `json:"count"`
	LastCompletedAt time.Time `json:"lastCompletedAt"`
}

func CreateDailyQuestProgress(id int) DailyQuestProgress {
	// set before the start of today so the user can complete the quest even if they signed up today
	lastCompletedAt := time.Now().Add(-time.Hour * 48).UTC().Round(time.Second)

	return DailyQuestProgress{DailyQuestID: id, LastCompletedAt: lastCompletedAt}
}

// Will check if the quest has already been completed today.
func (udq *DailyQuestProgress) IsCompleted() bool {
	y, m, d := time.Now().UTC().Date()
	startOfToday := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)

	return udq.LastCompletedAt.Unix() >= startOfToday.Unix()
}

// Will increase the count if the quest hasn't been completed today.
func (udq *DailyQuestProgress) IncreaseCount() {
	if !udq.IsCompleted() {
		udq.Count++
	}
}

// Call this method when completing the daily quest. It will update the state of the struct.
func (udq *DailyQuestProgress) Complete() {
	udq.Count = 0
	udq.LastCompletedAt = time.Now().UTC().Round(time.Second)
}
