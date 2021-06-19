package main

import "time"

type SignInRes struct {
	Token         string         `json:"token"`
	User          User           `json:"user"`
	UnitTemplates []UnitTemplate `json:"unitTemplates"`
}

type CampaignCollectRes struct {
	Exp             int       `json:"exp"`
	Gold            int       `json:"gold"`
	ExpStones       int       `json:"expStones"`
	LastCollectedAt time.Time `json:"lastCollectedAt"`
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
