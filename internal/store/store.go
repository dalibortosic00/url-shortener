package store

import (
	"context"

	"github.com/dalibortosic00/url-shortener/internal/models"
)

// Store defines the interface for URL shortener storage implementations.
// Different implementations can have different deduplication behaviors:
// - MemoryStore: deduplicates URLs (GetByURL returns existing codes)
// - DatabaseStore: allows multiple codes per URL for authorized users (GetByURL returns false)
type Store interface {
	// Save persists a link record. Returns ErrCollision if code already exists.
	Save(ctx context.Context, record *models.LinkRecord) error
	
	// Load retrieves the URL for a given code. Returns (url, true) if found.
	Load(ctx context.Context, code string) (string, bool)
	
	// GetByURL checks if a URL already has a code (for deduplication).
	// Returns (code, true) if found, ("", false) otherwise.
	// Database implementations should return ("", false) to allow multiple codes per URL.
	GetByURL(ctx context.Context, url string) (string, bool)
}
