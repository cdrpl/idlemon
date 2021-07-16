package main

import (
	"time"
)

const (
	VERSION           = "0.0.1"        // The current version of the server.
	ENV_FILE          = ".env"         // Default path to the .env file
	MAX_REQ_BODY_SIZE = 512            // Maximum number of bytes allowed in a request body.
	API_TOKEN_LEN     = 32             // Number of characters in the API token.
	API_TOKEN_TTL     = time.Hour * 12 // Time until the API token expires.
	MAX_PG_CONN       = 10             // Maximum number of open Postgres connections.
	UNIT_SUMMON_COST  = 250            // The cost to summon a unit.
	BCRYPT_COST       = 11             // The bcrypt cost used to hash a user's password.
	CHAT_LOG_LEN      = 15             // The amount of chat messages returned when fetching chat history.
)

const (
	CAMPAIGN_MAX_COLLECT       = time.Hour * 24 // Max time before campaign cannot collect anymore
	CAMPAIGN_EXP_PER_SEC       = 5              // The amount of exp earned every second on campaign level 1
	CAMPAIGN_GOLD_PER_SEC      = 20             // The amount of gold earned every second on campaign level 1
	CAMPAIGN_EXP_STONE_PER_SEC = 2              // The amount of exp stones earned every second on campaign level 1
	CAMPAIGN_EXP_GROWTH        = 2              // Exp gained from campaign increase by this value every 5 levels
	CAMPAIGN_GOLD_GROWTH       = 1              // Gold gained from campaign increase by this value every 5 levels
	CAMPAIGN_EXP_STONE_GROWTH  = 3              // Exp stones gained from campaign increase by this value every 5 levels
)

// Request DTOs validation.
const (
	CHAT_MESSAGE_MIN_LEN = 1
	CHAT_MESSAGE_MAX_LEN = 255
	USER_NAME_MIN        = 2
	USER_NAME_MAX        = 16
	USER_EMAIL_MAX       = 255
	USER_PASS_MIN        = 8
	USER_PASS_MAX        = 255
)

// Unit types, must have the same value as their table row IDs.
const (
	UNIT_TYPE_FOREST = iota
	UNIT_TYPE_ABYSS
	UNIT_TYPE_FORTRESS
	UNIT_TYPE_SHADOW
	UNIT_TYPE_LIGHT
	UNIT_TYPE_DARK
)

// Resources, must have the same value as their table row IDs.
const (
	RESOURCE_GOLD = iota
	RESOURCE_GEMS
	RESOURCE_EXP_STONE
	RESOURCE_EVO_STONE
)

// Daily quest IDs.
const (
	DAILY_QUEST_SIGN_IN = iota
)

// WebSocket message types.
const (
	WS_CHAT_MESSAGE = iota
)

// Transaction types.
const (
	TRANSACTION_GEMS = iota
	TRANSACTION_GOLD
	TRANSACTION_EXP_STONES
	TRANSACTION_USER_EXP
)

const (
	WS_WRITE_TIMOUT      = 10 * time.Second           // Time allowed to write a message to the peer.
	WS_PONG_TIMEOUT      = 60 * time.Second           // Pong must be received before this timout or else the connection will be closed.
	WS_PING_INTERVAL     = (WS_PONG_TIMEOUT * 9) / 10 // Send pings every interval. Must be less than pongWait.
	WS_MAX_MESSAGE_SIZE  = 512                        // Maximum message size allowed from peer.
	WS_READ_BUFFER_SIZE  = 1024
	WS_WRITE_BUFFER_SIZE = 1024
)

// Request context keys
const (
	UserIdCtx ctxKey = iota // The user ID of the authenticated user.
	ReqDtoCtx ctxKey = iota // Used for request DTOs.
)

type ctxKey int // Context key for adding data to the request context.
