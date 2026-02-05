package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dalibortosic00/url-shortener/internal/config"
	"github.com/dalibortosic00/url-shortener/internal/generator"
	"github.com/dalibortosic00/url-shortener/internal/handlers"
	"github.com/dalibortosic00/url-shortener/internal/services"
	"github.com/dalibortosic00/url-shortener/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	cfg := config.Load()

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
		MaxAge:         300,
	}))

	shortenerService := services.NewShortenerService(store.NewMemoryStore(), generator.NewRandomGenerator(6))
	shortenHandler := handlers.NewShortenHandler(shortenerService, cfg.BaseURL)
	resolveHandler := handlers.NewResolveHandler(shortenerService)

	router.Get("/health", handlers.Health)
	router.Post("/shorten", shortenHandler.Shorten)
	router.Get("/{code}", resolveHandler.Resolve)

	server := &http.Server{
		Handler:      router,
		Addr:         ":" + cfg.Port,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Starting server on port %s with BaseURL %s", cfg.Port, cfg.BaseURL)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
