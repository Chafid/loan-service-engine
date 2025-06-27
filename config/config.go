package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	JwtSecret string
)

func LoadEnv(envPath ...string) {
	path := ".env" // default
	if len(envPath) > 0 {
		path = envPath[0]
	}
	err := godotenv.Load(path)
	if err != nil {
		log.Fatal("Failed to load .env file")
	}
	JwtSecret = getEnv("JWT_SECRET", "")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
