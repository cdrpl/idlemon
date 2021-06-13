package main_test

import (
	"testing"
	"time"

	. "github.com/cdrpl/idlemon-server"
)

func TestCampaignCollect(t *testing.T) {
	timeDiff := 21
	now := time.Now()

	// represent campaign collected timeDiff seconds ago
	campaign := Campaign{
		Level:           13,
		LastCollectedAt: now.Add(-time.Second * time.Duration(timeDiff)),
	}

	exp, gold, expStones := campaign.Collect()

	if campaign.LastCollectedAt != time.Now().UTC().Round(time.Second) {
		t.Errorf("campaign did not set last collected at to now: %v", campaign.LastCollectedAt)
	}

	expect := timeDiff * (CAMPAIGN_EXP_PER_SEC + (campaign.Level / 5 * CAMPAIGN_EXP_GROWTH))
	if exp != expect {
		t.Errorf("invalid exp value, expect %v, receive %v", expect, exp)
	}

	expect = timeDiff * (CAMPAIGN_GOLD_PER_SEC + (campaign.Level / 5 * CAMPAIGN_GOLD_GROWTH))
	if gold != expect {
		t.Errorf("invalid gold value, expect %v, receive %v", expect, gold)
	}

	expect = timeDiff * (CAMPAIGN_EXP_STONE_PER_SEC + (campaign.Level / 5 * CAMPAIGN_EXP_STONE_GROWTH))
	if expStones != expect {
		t.Errorf("invalid exp stones value, expect %v, receive %v", expect, expStones)
	}
}
