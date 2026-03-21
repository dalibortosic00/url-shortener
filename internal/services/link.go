package services

import (
	"context"
	"errors"
	"time"

	"github.com/dalibortosic00/url-shortener/internal/models"
)

// LinkStore defines the interface for URL shortener storage implementations.
// - DatabaseStore: allows multiple codes per URL for authorized users (GetByURL returns false)
type LinkStore interface {
	// SaveLink persists a link record. Returns ErrCollision if code already exists.
	SaveLink(ctx context.Context, record *models.LinkRecord) error

	// LoadLink retrieves the URL for a given code. Returns (url, true) if found.
	LoadLink(ctx context.Context, code string) (*models.LinkRecord, bool)

	// GetCodeByURL checks if a URL already has a code (for deduplication).
	// Returns (code, true) if found, ("", false) otherwise.
	GetCodeByURL(ctx context.Context, url string) (string, bool)
}

type CodeGenerator interface {
	Generate(length int) (string, error)
}

type LinkService struct {
	store      LinkStore
	generator  CodeGenerator
	maxRetries int
}

func NewLinkService(store LinkStore, generator CodeGenerator) *LinkService {
	return &LinkService{
		store:      store,
		generator:  generator,
		maxRetries: 3,
	}
}

func (s *LinkService) Create(ctx context.Context, url string, ownerID *string) (string, error) {
	if ownerID == nil {
		if existingCode, exists := s.store.GetCodeByURL(ctx, url); exists {
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

		if err := s.store.SaveLink(ctx, record); err == nil {
			return code, nil
		} else if !errors.Is(err, models.ErrCollision) {
			return "", err
		}
	}

	return "", models.ErrFailedToGenerate
}

func (s *LinkService) CreateCustom(ctx context.Context, url string, customCode string, ownerID *string) (string, error) {
	record := &models.LinkRecord{
		Code:      customCode,
		URL:       url,
		OwnerID:   ownerID,
		CreatedAt: time.Now(),
	}

	if err := s.store.SaveLink(ctx, record); err != nil {
		return "", err
	}

	return customCode, nil
}

func (s *LinkService) Resolve(ctx context.Context, code string) (string, bool) {
	if record, ok := s.store.LoadLink(ctx, code); ok {
		return record.URL, true
	}

	return "", false
}
