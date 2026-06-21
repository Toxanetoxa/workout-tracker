package config

import "os"

type Config struct {
	HTTPAddr    string
	DatabaseURL string
}

func Load() Config {
	return Config{
		HTTPAddr:    getEnv("HTTP_ADDR", ":3000"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/workout_tracker?sslmode=disable"),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
