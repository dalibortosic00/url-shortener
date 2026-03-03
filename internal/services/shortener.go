package services

import (
	"context"
	"errors"
	"time"

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

type CodeGenerator interface {
	Generate(length int) (string, error)
}

type ShortenerService struct {
	publicStore  LinkStore
	privateStore LinkStore
	generator    CodeGenerator
	maxRetries   int
}

func NewShortenerService(publicStore LinkStore, privateStore LinkStore, generator CodeGenerator) *ShortenerService {
	return &ShortenerService{
		publicStore:  publicStore,
		privateStore: privateStore,
		generator:    generator,
		maxRetries:   3,
	}
}

func (s *ShortenerService) Create(ctx context.Context, url string, ownerID string) (string, error) {
	store := s.privateStore
	if ownerID == "" {
		store = s.publicStore
		if existingCode, exists := store.GetCodeByURL(ctx, url); exists {
			return existingCode, nil
		}
	}

	for i := 0; i < s.maxRetries; i++ {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		code, err := s.generator.Generate(6)
		if err != nil {
			return "", err
		}

		record := &models.LinkRecord{
			Code:      code,
			URL:       url,
			OwnerID:   ownerID,
			CreatedAt: time.Now(),
		}

		if err := store.SaveLink(ctx, record); err == nil {
			return code, nil
		} else if !errors.Is(err, models.ErrCollision) {
			return "", err
		}
	}

	return "", models.ErrFailedToGenerate
}

func (s *ShortenerService) Resolve(ctx context.Context, code string) (string, bool) {
	if record, ok := s.publicStore.LoadLink(ctx, code); ok {
		return record.URL, true
	}

	if record, ok := s.privateStore.LoadLink(ctx, code); ok {
		return record.URL, true
	}

	return "", false
}
