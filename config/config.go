package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DB_DSN      string
	PORT        string
	MAX_WORKERS string
	MAX_RETRIES string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		DB_DSN:      os.Getenv("DB_DSN"),
		PORT:        os.Getenv("PORT"),
		MAX_WORKERS: os.Getenv("MAX_WORKERS"),
		MAX_RETRIES: os.Getenv("MAX_RETRIES"),
	}
}
