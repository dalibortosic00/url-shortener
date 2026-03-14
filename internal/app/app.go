package app

import (
	"context"
	"database/sql"
	"fmt"
	"io"
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
	server  *http.Server
	db      *sql.DB
	logFile *os.File
}

func New(cfg *config.Config) (*App, error) {
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	multi := io.MultiWriter(os.Stdout, logFile)
	logger := log.New(multi, "", log.LstdFlags)
	log.SetOutput(multi)

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
	srv := server.New(cfg, userService, linkService, authMiddleware, logger)

	return &App{server: srv, db: db, logFile: logFile}, nil
}

func (a *App) Run() {
	defer func() {
		if err := a.db.Close(); err != nil {
			log.Printf("db close: %v", err)
		}
		if err := a.logFile.Close(); err != nil {
			log.Printf("log file close: %v", err)
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
