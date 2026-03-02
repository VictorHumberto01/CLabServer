package initializers

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnvVariables() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found. Accessing environment variables from system.")
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("CRITICAL ERROR: JWT_SECRET environment variable is empty. The server cannot start securely.")
	}
	if len(secret) < 32 {
		log.Fatal("CRITICAL ERROR: JWT_SECRET must be at least 32 characters long for secure HMAC signing.")
	}
}
