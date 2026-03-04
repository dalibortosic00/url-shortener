package main

import (
	"log"

	"github.com/dalibortosic00/url-shortener/internal/app"
	"github.com/dalibortosic00/url-shortener/internal/config"
)

func main() {
	cfg := config.Load()

	app, err := app.New(cfg)
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	app.Run()
}
