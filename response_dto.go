package main

import "time"

type SignInRes struct {
	Token              string               `json:"token"`
	User               User                 `json:"user"`
	Campaign           Campaign             `json:"campaign"`
	DailyQuestProgress []DailyQuestProgress `json:"dailyQuestProgress"`
	Resources          []Resource           `json:"resources"`
	Units              []Unit               `json:"units"`
	UnitTemplates      []UnitTemplate       `json:"unitTemplates"`
}

type CampaignCollectRes struct {
	Transactions    [3]Transaction `json:"transactions"`
	LastCollectedAt time.Time      `json:"lastCollectedAt"`
}

type DailyQuestCompleteRes struct {
	Status      int         `json:"status"`
	Message     string      `json:"message"`
	Transaction Transaction `json:"transaction"`
}

type SummonUnitRes struct {
	Unit        Unit        `json:"unit"`
	Transaction Transaction `json:"transaction"`
}
