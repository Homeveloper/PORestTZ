package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort string
	DSN      string
}

func Load(path string) (*Config, error) {
	_ = godotenv.Load(path)

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	return &Config{
		HTTPPort: port,
		DSN:      dsn,
	}, nil
}