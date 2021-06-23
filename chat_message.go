package main

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

type ChatMessage struct {
	Id         int    `json:"id"`
	SenderName string `json:"senderName"`
	Message    string `json:"message"`
}

// Returns n amount of chat messages sent before start. Start is the ID of the chat message to start from.
// The next n chat messages ordered by ID will be returned. Set start to 0 to fetch the most recent messages.
func GetChatMessages(ctx context.Context, db *pgxpool.Pool, start int, num int) ([]ChatMessage, error) {
	chatMessages := make([]ChatMessage, 0)

	// if start is 0 then return the most recent messages
	if start == 0 {
		start = math.MaxInt32
	}

	query := `SELECT chat_messages.id, chat_messages.message, users.name FROM chat_messages
			  LEFT JOIN users ON chat_messages.user_id = users.id
			  WHERE chat_messages.id < $1
			  ORDER BY chat_messages.id DESC
			  LIMIT $2`

	rows, err := db.Query(ctx, query, start, num)
	if err != nil {
		return chatMessages, err
	}

	for rows.Next() {
		var cm ChatMessage

		if err := rows.Scan(&cm.Id, &cm.Message, &cm.SenderName); err != nil {
			return chatMessages, err
		}

		chatMessages = append(chatMessages, cm)
	}

	return chatMessages, nil
}

func InsertChatMessage(ctx context.Context, db *pgxpool.Pool, userId uuid.UUID, message string) error {
	query := "INSERT INTO chat_messages (user_id, message, sent_at) VALUES ($1, $2, $3)"
	_, err := db.Exec(ctx, query, userId, message, time.Now())

	return err
}
