package main

import "log"

// Represents resources given as a reward.
type Reward struct {
	Type   string `json:"type"`
	Amount int    `json:"amount"`
}

// Will apply the reward to a user.
func (r Reward) Apply(user *User) {
	switch r.Type {
	case GEMS:
		user.Data.Resources[RESOURCE_GEMS].Amount += r.Amount

	default:
		log.Fatalf("failed to apply reward of type %v, not handled in switch statement\n", r.Type)
	}
}
