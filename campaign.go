package main

import (
	"time"
)

type Campaign struct {
	Level           int       `json:"level"`
	LastCollectedAt time.Time `json:"lastCollectedAt"`
}

// Returns the amount of resources that can be collected and sets last collected at to now.
func (c *Campaign) Collect(user *User) (exp int, gold int, expStones int) {
	timeDiff := time.Since(c.LastCollectedAt)
	timeDiffSec := int(timeDiff.Seconds())

	// limit stockpile to 24 hours
	if timeDiff > CAMPAIGN_MAX_COLLECT {
		timeDiff = CAMPAIGN_MAX_COLLECT
	}

	if timeDiff > time.Second {
		exp = timeDiffSec * (CAMPAIGN_EXP_PER_SEC + (c.Level / 5 * CAMPAIGN_EXP_GROWTH))
		gold = timeDiffSec * (CAMPAIGN_GOLD_PER_SEC + (c.Level / 5 * CAMPAIGN_GOLD_GROWTH))
		expStones = timeDiffSec * (CAMPAIGN_EXP_STONE_PER_SEC + (c.Level / 5 * CAMPAIGN_EXP_STONE_GROWTH))
		c.LastCollectedAt = time.Now().UTC().Round(time.Second)

		user.Data.Exp += exp
		user.Data.Resources[RESOURCE_GOLD].Amount += gold
		user.Data.Resources[RESOURCE_EXP_STONE].Amount += expStones
	}

	return
}
