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

	CheckEnvVars()
}

func CheckEnvVars() {
	CheckEnvVar("ENV")
	CheckEnvVar("PORT")
	CheckEnvVar("CLIENT_VERSION")
	CheckEnvVar("DB_USER")
	CheckEnvVar("DB_PASS")
	CheckEnvVar("DB_NAME")
	CheckEnvVar("DB_HOST")
	CheckEnvVar("REDIS_HOST")
	CheckEnvVar("CREATE_TABLES")
	CheckEnvVar("DROP_TABLES")
	CheckEnvVar("ADMIN_NAME")
	CheckEnvVar("ADMIN_EMAIL")
	CheckEnvVar("ADMIN_PASS")
	CheckEnvVar("INSERT_ADMIN")
}

// Will log fatal if env var is not set.
func CheckEnvVar(key string) {
	if os.Getenv(key) == "" {
		log.Fatalf("environment variable %v must be set", key)
	}
}
