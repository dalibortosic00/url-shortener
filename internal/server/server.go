package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/dalibortosic00/url-shortener/internal/config"
	"github.com/dalibortosic00/url-shortener/internal/handlers"
	"github.com/dalibortosic00/url-shortener/internal/middleware"
	"github.com/dalibortosic00/url-shortener/internal/models"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type LinkService interface {
	Create(ctx context.Context, url string, ownerID *string) (string, error)
	CreateCustom(ctx context.Context, url string, customCode string, ownerID *string) (string, error)
	Resolve(ctx context.Context, code string) (string, bool)
	List(ctx context.Context, ownerID string) ([]models.LinkRecord, error)
	Delete(ctx context.Context, code string, ownerID string) error
}

type UserService interface {
	Create(ctx context.Context, name string) (string, error)
}

func New(
	cfg *config.Config,
	userService UserService,
	linkService LinkService,
	auth *middleware.AuthMiddleware,
	logger *log.Logger,
) *http.Server {
	r := chi.NewRouter()
	registerMiddleware(r, logger)
	registerRoutes(r, cfg, userService, linkService, auth)

	return &http.Server{
		Handler:      r,
		Addr:         ":" + cfg.Port,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

func registerMiddleware(r *chi.Mux, logger *log.Logger) {
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.GetHead)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
		MaxAge:         300,
	}))

	if logger != nil {
		r.Use(chiMiddleware.RequestLogger(&chiMiddleware.DefaultLogFormatter{
			Logger:  logger,
			NoColor: true,
		}))
	}
}

func registerRoutes(
	r *chi.Mux,
	cfg *config.Config,
	userService UserService,
	linkService LinkService,
	auth *middleware.AuthMiddleware,
) {
	shortenHandler := handlers.NewShortenHandler(linkService, cfg.BaseURL)
	linksHandler := handlers.NewLinksHandler(linkService, cfg.BaseURL)
	deleteHandler := handlers.NewDeleteHandler(linkService)
	resolveHandler := handlers.NewResolveHandler(linkService)
	registerHandler := handlers.NewRegisterHandler(userService)

	r.Post("/register", registerHandler.Register)
	r.Get("/{code}", resolveHandler.Resolve)

	r.Group(func(r chi.Router) {
		r.Use(auth.OptionalAuth)
		r.Post("/shorten", shortenHandler.Shorten)
	})

	r.Group(func(r chi.Router) {
		r.Use(auth.RequireAuth)
		r.Get("/links", linksHandler.List)
		r.Delete("/{code}", deleteHandler.Delete)
	})
}
