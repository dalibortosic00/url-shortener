package store

import (
	"context"

	"github.com/dalibortosic00/url-shortener/internal/models"
)

// LinkStore defines the interface for URL shortener storage implementations.
// Different implementations can have different deduplication behaviors:
// - MemoryStore: deduplicates URLs (GetByURL returns existing codes)
// - DatabaseStore: allows multiple codes per URL for authorized users (GetByURL returns false)
type LinkStore interface {
	// SaveLink persists a link record. Returns ErrCollision if code already exists.
	SaveLink(ctx context.Context, record *models.LinkRecord) error

	// LoadLink retrieves the URL for a given code. Returns (url, true) if found.
	LoadLink(ctx context.Context, code string) (*models.LinkRecord, bool)

	// GetCodeByURL checks if a URL already has a code (for deduplication).
	// Returns (code, true) if found, ("", false) otherwise.
	// Database implementations should return ("", false) to allow multiple codes per URL.
	GetCodeByURL(ctx context.Context, url string) (string, bool)
}

// UserStore defines the interface for user storage, allowing retrieval of users by API key.
type UserStore interface {
	// Save persists a user record.
	SaveUser(ctx context.Context, user *models.User) error

	// GetUserByAPIKey retrieves a user by their API key.
	// Returns (user, nil) if found, (nil, ErrRecordNotFound) if not found, or (nil, error) on other errors.
	GetUserByAPIKey(ctx context.Context, apiKey string) (*models.User, error)
}
