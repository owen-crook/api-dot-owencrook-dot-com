package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	GCPProjectID        string
	FirestoreDatabaseID string
	GeminiToken         string
	Environment         string
	GoogleClientID      string
	GoogleClientSecret  string
	AdminEmails         string
}

// LoadConfig reads environment variables into a Config struct.
func LoadConfig() *Config {

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, continuing with existing env vars")
	}

	cfg := &Config{
		AdminEmails:         getEnv("ADMIN_EMAILS", ""),
		GCPProjectID:        getEnv("GCP_PROJECT_ID", ""),
		FirestoreDatabaseID: getEnv("FIRESTORE_DATABASE_ID", ""),
		GeminiToken:         getEnv("GEMINI_API_KEY", ""),
		Environment:         getEnv("ENVIRONMENT", "LOCAL"),
		GoogleClientID:      getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:  getEnv("GOOGLE_CLIENT_SECRET", ""),
	}

	if cfg.AdminEmails == "" {
		log.Fatal("ADMIN_EMAILS is required")
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

	if cfg.GoogleClientSecret == "" {
		log.Fatal("GOOGLE_CLIENT_ID is required")
	}

	if cfg.FirestoreDatabaseID == "" {
		log.Fatal("GOOGLE_CLIENT_SECRET is required")
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

func GetAdminEmails(cfg Config) []string {
	emails := cfg.AdminEmails
	if emails == "" {
		return nil
	}
	return strings.Split(emails, ",")
}
