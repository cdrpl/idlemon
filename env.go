package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Call this function to load the .env file and set the default values.
func LoadEnv(path string, serverVersion string) {
	os.Setenv("SERVER_VERSION", serverVersion)

	if path != "nil" {
		if err := godotenv.Load(path); err != nil {
			log.Fatalln("failed to load .env file:", err)
		}
	}

	SetEnvDefaults()
}

func SetEnvDefaults() {
	SetEnvDefault("ENV", ENV)
	SetEnvDefault("PORT", PORT)
	SetEnvDefault("CLIENT_VERSION", CLIENT_VERSION)
	SetEnvDefault("ADMIN_PASS", ADMIN_PASS)
	SetEnvDefault("DB_USER", DB_USER)
	SetEnvDefault("DB_PASS", DB_PASS)
	SetEnvDefault("DB_NAME", DB_NAME)
	SetEnvDefault("DB_HOST", DB_HOST)
	SetEnvDefault("REDIS_HOST", REDIS_HOST)
	SetEnvDefault("RUN_MIGRATIONS", RUN_MIGRATIONS)
}

// set env var to default value if not set.
func SetEnvDefault(key string, val string) {
	if os.Getenv(key) == "" {
		if err := os.Setenv(key, val); err != nil {
			log.Fatalln(err)
		}
	}
}
