package helpers

import (
	"net/http/httptest"
	"testing"

	"github.com/dalibortosic00/url-shortener/internal/config"
	"github.com/dalibortosic00/url-shortener/internal/generators"
	"github.com/dalibortosic00/url-shortener/internal/middleware"
	"github.com/dalibortosic00/url-shortener/internal/server"
	"github.com/dalibortosic00/url-shortener/internal/services"
	"github.com/dalibortosic00/url-shortener/internal/store"
)

func NewTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	cfg := config.Load()
	if cfg.TestDatabaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping e2e tests")
	}

	db := NewTestDB(t, cfg.TestDatabaseURL)
	gen := generators.NewRandomGenerator()

	publicStore := store.NewMemoryStore()
	privateStore := store.NewDatabaseStore(db)

	userService := services.NewUserService(privateStore, gen)
	linkService := services.NewLinkService(publicStore, privateStore, gen)
	authMiddleware := middleware.NewAuthMiddleware(userService)

	srv := server.New(cfg, userService, linkService, authMiddleware)

	ts := httptest.NewServer(srv.Handler)
	t.Cleanup(func() {
		Reset(t, db)
		ts.Close()
	})

	return ts
}
