package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Daily quest has a required number to complete it and a transaction representing the reward given for completion.
type DailyQuest struct {
	Id          int
	Required    int
	Transaction Transaction
}

// Returns a slice of all the daily quests
func DailyQuests() []DailyQuest {
	dailyQuests := make([]DailyQuest, 1)

	// require user to sign in
	dailyQuests[DAILY_QUEST_SIGN_IN] = DailyQuest{
		Id:       DAILY_QUEST_SIGN_IN,
		Required: 1,
		Transaction: Transaction{
			Type:   TRANSACTION_GEMS,
			Amount: 1000,
		},
	}

	return dailyQuests
}

type DailyQuestProgress struct {
	Id              int       `json:"id"`
	UserId          uuid.UUID `json:"-"`
	DailyQuestId    int       `json:"dailyQuestId"`
	Count           int       `json:"count"`
	LastCompletedAt time.Time `json:"lastCompletedAt"`
}

func CreateDailyQuestProgress() DailyQuestProgress {
	// set before the start of today so the user can complete the quest even if they just signed up
	lastCompletedAt := time.Now().Add(-time.Hour * 48)

	return DailyQuestProgress{LastCompletedAt: lastCompletedAt}
}

// Will check if the quest has already been completed today.
func (dqp *DailyQuestProgress) IsCompleted() bool {
	now := time.Now()
	y, m, d := now.Date()
	startOfToday := time.Date(y, m, d, 0, 0, 0, 0, now.Location())

	return dqp.LastCompletedAt.Unix() >= startOfToday.Unix()
}

// Will increase the count if the quest hasn't been completed today.
func (dqp *DailyQuestProgress) IncreaseCount(ctx context.Context, tx pgx.Tx) error {
	if !dqp.IsCompleted() {
		_, err := tx.Exec(ctx, "UPDATE daily_quest_progress SET count = count + 1 WHERE id = $1", dqp.Id)
		if err != nil {
			return fmt.Errorf("fail to update daily_quest_progress table: %w", err)
		}

		dqp.Count++
	}

	return nil
}

// Call this method when completing the daily quest. It will update the state of the struct.
func (dqp *DailyQuestProgress) Complete(ctx context.Context, tx pgx.Tx) error {
	completedAt := time.Now()

	query := "UPDATE daily_quest_progress SET count = 0, last_completed_at = $1 WHERE id = $2"
	_, err := tx.Exec(ctx, query, completedAt, dqp.Id)
	if err != nil {
		return fmt.Errorf("fail to update daily_quest_progress table: %w", err)
	}

	dqp.Count = 0
	dqp.LastCompletedAt = completedAt

	return nil
}

func FindAllDailyQuestProgress(ctx context.Context, db *pgxpool.Pool, userId uuid.UUID) ([]DailyQuestProgress, error) {
	dailyQuestProgress := make([]DailyQuestProgress, 0)

	query := "SELECT id, daily_quest_id, count, last_completed_at FROM daily_quest_progress WHERE user_id = $1"
	rows, err := db.Query(ctx, query, userId)
	if err != nil {
		return dailyQuestProgress, fmt.Errorf("faily to query daily_quest_progress table: %w", err)
	}

	for rows.Next() {
		var progress DailyQuestProgress

		err := rows.Scan(&progress.Id, &progress.DailyQuestId, &progress.Count, &progress.LastCompletedAt)
		if err != nil {
			return dailyQuestProgress, fmt.Errorf("fail to scan rows: %w", err)
		}

		dailyQuestProgress = append(dailyQuestProgress, progress)
	}

	return dailyQuestProgress, nil
}

func FindDailyQuestProgress(ctx context.Context, tx pgx.Tx, userId uuid.UUID, dailyQuestId int) (DailyQuestProgress, error) {
	progress := DailyQuestProgress{UserId: userId, DailyQuestId: dailyQuestId}

	query := "SELECT id, count, last_completed_at FROM daily_quest_progress WHERE (user_id = $1 AND daily_quest_id = $2)"
	err := tx.QueryRow(ctx, query, userId, dailyQuestId).Scan(&progress.Id, &progress.Count, &progress.LastCompletedAt)
	if err != nil {
		return progress, fmt.Errorf("fail to query daily_quest_progress table: %w", err)
	}

	return progress, nil
}

func InsertDailyQuestProgress(ctx context.Context, tx pgx.Tx, dc *DataCache, userId uuid.UUID) error {
	for _, dailyQuest := range dc.DailyQuests {
		progress := CreateDailyQuestProgress()

		query := "INSERT INTO daily_quest_progress (user_id, daily_quest_id, last_completed_at) VALUES ($1, $2, $3)"
		_, err := tx.Exec(ctx, query, userId, dailyQuest.Id, progress.LastCompletedAt)
		if err != nil {
			return fmt.Errorf("fail to insert daily quest progress: %w", err)
		}
	}

	return nil
}
