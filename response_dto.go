package main

import "time"

// Sign in response
type SignInRes struct {
	Token         string         `json:"token"`
	User          User           `json:"user"`
	Units         []Unit         `json:"units"`
	UserResources []UserResource `json:"userResources"`
	Campaign      Campaign       `json:"campaign"`
	Resources     []Resource     `json:"resources"`
}

type CampaignCollectRes struct {
	Exp             int       `json:"exp"`
	Gold            int       `json:"gold"`
	ExpStones       int       `json:"expStones"`
	LastCollectedAt time.Time `json:"lastCollectedAt"`
}
