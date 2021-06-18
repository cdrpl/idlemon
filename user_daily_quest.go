package main

import (
	"log"
)

type UserDailyQuest struct {
	Count       int `json:"count"`
	IsCompleted int `json:"isCompleted"`
}

func CompleteDailyQuest(id int) Reward {
	switch id {
	case DAILY_QUEST_SIGN_IN:
		return Reward{Type: GEMS, Amount: DAILY_SIGN_IN_REWARD}

	default:
		log.Printf("attempt to complete daily quest not handled by switch statement: %v\n", id)
	}

	return Reward{}
}

func CompleteSignInDailyQuest() {

}
