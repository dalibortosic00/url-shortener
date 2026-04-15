package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                  string
	BaseURL               string
	DatabaseURL           string
	TestDatabaseURL       string
	RateLimitAnonRequests int
	RateLimitAnonWindow   time.Duration
	RateLimitAuthRequests int
	RateLimitAuthWindow   time.Duration
}

func Load() *Config {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is required")
	}

	baseURL := strings.TrimRight(os.Getenv("BASE_URL"), "/")
	if baseURL == "" {
		baseURL = "http://localhost:" + port
	}

	databaseURL := os.Getenv("DB_URL")
	if databaseURL == "" {
		log.Fatal("DB_URL environment variable is required")
	}

	testDatabaseURL := os.Getenv("TEST_DB_URL")

	rateLimitAnonRequests := 5
	if val := os.Getenv("RATE_LIMIT_ANON_REQUESTS"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			rateLimitAnonRequests = parsed
		}
	}

	rateLimitAnonWindow := 24 * time.Hour
	if val := os.Getenv("RATE_LIMIT_ANON_WINDOW"); val != "" {
		if parsed, err := time.ParseDuration(val); err == nil {
			rateLimitAnonWindow = parsed
		}
	}

	rateLimitAuthRequests := 5
	if val := os.Getenv("RATE_LIMIT_AUTH_REQUESTS"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			rateLimitAuthRequests = parsed
		}
	}

	rateLimitAuthWindow := 1 * time.Hour
	if val := os.Getenv("RATE_LIMIT_AUTH_WINDOW"); val != "" {
		if parsed, err := time.ParseDuration(val); err == nil {
			rateLimitAuthWindow = parsed
		}
	}

	return &Config{
		Port:                  port,
		BaseURL:               baseURL,
		DatabaseURL:           databaseURL,
		TestDatabaseURL:       testDatabaseURL,
		RateLimitAnonRequests: rateLimitAnonRequests,
		RateLimitAnonWindow:   rateLimitAnonWindow,
		RateLimitAuthRequests: rateLimitAuthRequests,
		RateLimitAuthWindow:   rateLimitAuthWindow,
	}
}
