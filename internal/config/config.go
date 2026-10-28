package config

import "os"

const defaultDatabaseURL = "postgres://biblioteca:biblioteca@localhost:5432/biblioteca?sslmode=disable"

type Config struct {
	DatabaseURL string
}

func Load() Config {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = defaultDatabaseURL
	}
	return Config{DatabaseURL: url}
}
