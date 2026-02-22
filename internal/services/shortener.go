package services

import (
	"context"
	"errors"
	"time"

	"github.com/dalibortosic00/url-shortener/internal/generators"
	"github.com/dalibortosic00/url-shortener/internal/models"
	"github.com/dalibortosic00/url-shortener/internal/store"
)

type ShortenerService struct {
	publicStore  store.LinkStore
	privateStore store.LinkStore
	generator    *generators.RandomGenerator
	maxRetries   int
}

func NewShortenerService(publicStore store.LinkStore, privateStore store.LinkStore, generator *generators.RandomGenerator) *ShortenerService {
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
