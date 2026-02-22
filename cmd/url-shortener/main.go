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
	"github.com/dalibortosic00/url-shortener/internal/generators"
	"github.com/dalibortosic00/url-shortener/internal/handlers"
	"github.com/dalibortosic00/url-shortener/internal/middleware"
	"github.com/dalibortosic00/url-shortener/internal/services"
	"github.com/dalibortosic00/url-shortener/internal/store"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	cfg := config.Load()

	router := chi.NewRouter()

	router.Use(chiMiddleware.RequestID)
	router.Use(chiMiddleware.Logger)
	router.Use(chiMiddleware.Recoverer)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
		MaxAge:         300,
	}))

	db, err := store.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	publicStore := store.NewMemoryStore()
	privateStore := store.NewDatabaseStore(db)

	authMiddleware := middleware.NewAuthMiddleware(privateStore)
	randomGenerator := generators.NewRandomGenerator()
	shortenerService := services.NewShortenerService(publicStore, privateStore, randomGenerator)
	shortenHandler := handlers.NewShortenHandler(shortenerService, cfg.BaseURL)
	resolveHandler := handlers.NewResolveHandler(shortenerService)

	userService := services.NewUserService(privateStore, randomGenerator)
	registerHandler := handlers.NewRegisterHandler(userService)

	router.Get("/health", handlers.Health)
	router.Post("/shorten", authMiddleware.Middleware(shortenHandler.Shorten))
	router.Get("/{code}", resolveHandler.Resolve)
	router.Post("/register", registerHandler.Register)

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
