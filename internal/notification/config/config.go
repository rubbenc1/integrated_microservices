package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	KAFKA_BROKER string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	return &Config{
		KAFKA_BROKER: os.Getenv("KAFKA_BROKER"),
	}
}
