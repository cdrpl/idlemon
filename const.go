package main

import "time"

const (
	VERSION           = "0.0.1"        // The current version of the server.
	ENV_FILE          = ".env"         // Path to the .env file
	DB_CONN_RETRIES   = 6              // Number of database connection retries before exit
	MAX_REQ_BODY_SIZE = 512            // Maximum number of bytes allowed in a request body.
	API_TOKEN_LEN     = 32             // Number of characters in the API token.
	API_TOKEN_TTL     = time.Hour * 12 // Time until the API token expires.
)

// Default env values.
const (
	ENV            = "development"
	PORT           = "3000"
	CLIENT_VERSION = "1.0.0"
	ADMIN_PASS     = "adminpass"
	DB_USER        = "root"
	DB_PASS        = "password"
	DB_NAME        = "idlemon"
	DB_HOST        = "localhost"
	REDIS_HOST     = "localhost"
	RUN_MIGRATIONS = "true"
)

// Unit types, must have the same value as their table row IDs.
const (
	UNIT_TYPE_FOREST = iota + 1
	UNIT_TYPE_ABYSS
	UNIT_TYPE_FORTRESS
	UNIT_TYPE_SHADOW
	UNIT_TYPE_LIGHT
	UNIT_TYPE_DARK
)

// Resources, must have the same value as their table row IDs.
const (
	RESOURCE_GOLD = iota + 1
	RESOURCE_GEMS
	RESOURCE_EXP_STONE
	RESOURCE_EVO_STONE
)

// MariaDB error codes
const (
	ER_DUP_ENTRY = 1062
)

// Admin user
const (
	ADMIN_ID    = 1
	ADMIN_NAME  = "Admin"
	ADMIN_EMAIL = "admin@idlemon.com"
)

const (
	WS_WRITE_TIMOUT      = 10 * time.Second           // Time allowed to write a message to the peer.
	WS_PONG_TIMEOUT      = 60 * time.Second           // Time allowed to read the next pong message from the peer.
	WS_PING_PERIOD       = (WS_PONG_TIMEOUT * 9) / 10 // Send pings to peer with this period. Must be less than pongWait.
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
