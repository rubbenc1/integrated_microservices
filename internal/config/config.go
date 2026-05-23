package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost            string
	DBPort            string
	POSTGRES_DB       string
	POSTGRES_USER     string
	POSTGRES_PASSWORD string
	JWT_SECRET        string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	return &Config{
		DBHost:            os.Getenv("DATABASE_HOST"),
		DBPort:            os.Getenv("DATABASE_PORT"),
		POSTGRES_DB:       os.Getenv("POSTGRES_DB"),
		POSTGRES_USER:     os.Getenv("POSTGRES_USER"),
		POSTGRES_PASSWORD: os.Getenv("POSTGRES_PASSWORD"),
		JWT_SECRET:        os.Getenv("JWT_SECRET"),
	}
}
