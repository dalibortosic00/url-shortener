package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	BaseURL         string
	DatabaseURL     string
	TestDatabaseURL string
}

func Load() *Config {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is required")
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:" + port
	}

	databaseURL := os.Getenv("DB_URL")
	if databaseURL == "" {
		log.Fatal("DB_URL environment variable is required")
	}

	testDatabaseURL := os.Getenv("TEST_DB_URL")

	return &Config{
		Port:            port,
		BaseURL:         baseURL,
		DatabaseURL:     databaseURL,
		TestDatabaseURL: testDatabaseURL,
	}
}
