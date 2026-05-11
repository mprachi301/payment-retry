package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DB_DSN      string
	PORT        string
	MAX_WORKERS int
	MAX_RETRIES int
}

func Load() *Config {
	godotenv.Load()

	maxWorkers, _ := strconv.Atoi(os.Getenv("MAX_WORKERS"))
	maxRetries, _ := strconv.Atoi(os.Getenv("MAX_RETRIES"))

	return &Config{
		DB_DSN:      os.Getenv("DB_DSN"),
		PORT:        os.Getenv("PORT"),
		MAX_WORKERS: maxWorkers,
		MAX_RETRIES: maxRetries,
	}
}
