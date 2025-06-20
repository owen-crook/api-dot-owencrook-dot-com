package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	GCPProjectID       string
	GCPCredentialsFile string
	GeminiToken        string
	HuggingFaceToken   string
	Environment        string
}

// LoadConfig reads environment variables into a Config struct.
func LoadConfig() *Config {

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, continuing with existing env vars")
	}

	cfg := &Config{
		Port:               getEnv("PORT", "8080"),
		GCPProjectID:       getEnv("GCP_PROJECT_ID", ""),
		GCPCredentialsFile: getEnv("GOOGLE_APPLICATION_CREDENTIALS", ""),
		GeminiToken:        getEnv("GEMINI_API_KEY", ""),
		HuggingFaceToken:   getEnv("HUGGING_FACE_INFERENCE_TOKEN", ""),
		Environment:        getEnv("ENVIRONMENT", "development"),
	}

	if cfg.GCPProjectID == "" {
		log.Fatal("GCP_PROJECT_ID is required")
	}

	if cfg.GCPCredentialsFile == "" {
		log.Fatal("GOOGLE_APPLICATION_CREDENTIALS is required")
	}

	if cfg.GeminiToken == "" {
		log.Fatal("GEMINI_API_KEY is required")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
