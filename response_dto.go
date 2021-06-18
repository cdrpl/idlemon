package main

import "time"

// Sign in response
type SignInRes struct {
	Token         string         `json:"token"`
	User          User           `json:"user"`
	DailyQuests   []DailyQuest   `json:"dailyQuests"`
	Resources     []Resource     `json:"resources"`
	UnitTemplates []UnitTemplate `json:"unitTemplates"`
	UnitType      []UnitType     `json:"unitTypes"`
}

type CampaignCollectRes struct {
	Exp             int       `json:"exp"`
	Gold            int       `json:"gold"`
	ExpStones       int       `json:"expStones"`
	LastCollectedAt time.Time `json:"lastCollectedAt"`
}

type DailyQuestCompleteRes struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Reward  Reward `json:"reward"`
}
