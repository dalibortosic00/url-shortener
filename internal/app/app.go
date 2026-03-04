package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dalibortosic00/url-shortener/internal/config"
	"github.com/dalibortosic00/url-shortener/internal/generators"
	"github.com/dalibortosic00/url-shortener/internal/middleware"
	"github.com/dalibortosic00/url-shortener/internal/server"
	"github.com/dalibortosic00/url-shortener/internal/services"
	"github.com/dalibortosic00/url-shortener/internal/store"
)

type App struct {
	server *http.Server
	db     *sql.DB
}

func New(cfg *config.Config) (*App, error) {
	db, err := store.InitDB(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("init db: %w", err)
	}

	publicStore := store.NewMemoryStore()
	privateStore := store.NewDatabaseStore(db)

	randomGenerator := generators.NewRandomGenerator()

	userService := services.NewUserService(privateStore, randomGenerator)
	linkService := services.NewLinkService(publicStore, privateStore, randomGenerator)

	authMiddleware := middleware.NewAuthMiddleware(userService)
	srv := server.New(cfg, userService, linkService, authMiddleware)

	return &App{server: srv, db: db}, nil
}

func (a *App) Run() {
	defer func() {
		if err := a.db.Close(); err != nil {
			log.Printf("db close: %v", err)
		}
	}()

	go func() {
		log.Printf("starting on %s", a.server.Addr)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}

	log.Println("exited cleanly")
}
