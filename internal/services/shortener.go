package services

import (
	"context"
	"errors"
	"time"

	"github.com/dalibortosic00/url-shortener/internal/generator"
	"github.com/dalibortosic00/url-shortener/internal/models"
	"github.com/dalibortosic00/url-shortener/internal/store"
)

type ShortenerService struct {
	publicStore  store.Store
	privateStore store.Store
	generator    *generator.RandomGenerator
	maxRetries   int
}

func NewShortenerService(publicStore store.Store, privateStore store.Store, generator *generator.RandomGenerator) *ShortenerService {
	return &ShortenerService{
		publicStore:  publicStore,
		privateStore: privateStore,
		generator:    generator,
		maxRetries:   3,
	}
}

// TODO: Handle private store for authenticated users
func (s *ShortenerService) Create(ctx context.Context, url string) (string, error) {
	if existingCode, exists := s.publicStore.GetByURL(ctx, url); exists {
		return existingCode, nil
	}

	for i := 0; i < s.maxRetries; i++ {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		code := s.generator.Generate()

		record := &models.LinkRecord{
			Code:      code,
			URL:       url,
			CreatedAt: time.Now(),
		}

		err := s.publicStore.Save(ctx, record)
		if err == nil {
			return code, nil
		}

		if !errors.Is(err, models.ErrCollision) {
			return "", err
		}
	}

	return "", models.ErrFailedToGenerate
}

// TODO: Handle private store for authenticated users
func (s *ShortenerService) Resolve(ctx context.Context, code string) (string, bool) {
	select {
	case <-ctx.Done():
		return "", false
	default:
	}

	return s.publicStore.Load(ctx, code)
}
