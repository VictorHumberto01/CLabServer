package initializers

import (
	"log"

	"github.com/joho/godotenv"
)

func LoadEnvVariables() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found. Accessing environment variables from system.")
	}
}
