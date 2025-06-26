package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	GCPProjectID        string
	FirestoreDatabaseID string
	GeminiToken         string
	Environment         string
}

// LoadConfig reads environment variables into a Config struct.
func LoadConfig() *Config {

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, continuing with existing env vars")
	}

	cfg := &Config{
		GCPProjectID:        getEnv("GCP_PROJECT_ID", ""),
		FirestoreDatabaseID: getEnv("FIRESTORE_DATABASE_ID", ""),
		GeminiToken:         getEnv("GEMINI_API_KEY", ""),
		Environment:         getEnv("ENVIRONMENT", "LOCAL"),
	}

	if cfg.GCPProjectID == "" {
		log.Fatal("GCP_PROJECT_ID is required")
	}

	if cfg.FirestoreDatabaseID == "" {
		log.Fatal("FirestoreDatabaseID is required")
	}

	if cfg.GeminiToken == "" {
		log.Fatal("GEMINI_API_KEY is required")
	}

	if cfg.Environment == "LOCAL" && getEnv("GOOGLE_APPLICATION_CREDENTIALS", "") == "" {
		log.Fatal("GOOGLE_APPLICATION_CREDENTIALS required for local development")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
