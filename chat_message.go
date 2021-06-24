package main

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

type ChatMessage struct {
	Id         int       `json:"id"`
	SenderName string    `jso-n:"senderName"`
	Message    string    `json:"message"`
	SentAt     time.Time `json:"sentAt"`
}

// Returns n amount of chat messages sent before start. Start is the ID of the chat message to start from.
// The next n chat messages ordered by ID will be returned. Set start to 0 to fetch the most recent messages.
func GetChatMessages(ctx context.Context, db *pgxpool.Pool, start int, num int) ([]ChatMessage, error) {
	chatMessages := make([]ChatMessage, 0)

	// if start is 0 then return the most recent messages
	if start == 0 {
		start = math.MaxInt32
	}

	query := `SELECT id, sender_name, message, sent_at FROM chat_messages
			  WHERE id < $1
			  ORDER BY id DESC
			  LIMIT $2`

	rows, err := db.Query(ctx, query, start, num)
	if err != nil {
		return chatMessages, err
	}

	for rows.Next() {
		var cm ChatMessage

		if err := rows.Scan(&cm.Id, &cm.SenderName, &cm.Message, &cm.SentAt); err != nil {
			return chatMessages, err
		}

		chatMessages = append(chatMessages, cm)
	}

	return chatMessages, nil
}

// Insert chat message into the database and return the ID.
func InsertChatMessage(ctx context.Context, db *pgxpool.Pool, senderId uuid.UUID, senderName string, message string) (int, error) {
	var id int

	query := "INSERT INTO chat_messages (user_id, sender_name, message, sent_at) VALUES ($1, $2, $3, $4) RETURNING id"
	err := db.QueryRow(ctx, query, senderId, senderName, message, time.Now()).Scan(&id)

	return id, err
}
