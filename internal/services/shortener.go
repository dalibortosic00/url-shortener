package services

import (
	"context"
	"errors"

	"github.com/dalibortosic00/url-shortener/internal/generator"
	"github.com/dalibortosic00/url-shortener/internal/models"
	"github.com/dalibortosic00/url-shortener/internal/store"
)

type ShortenerService struct {
	store      *store.MemoryStore
	generator  *generator.RandomGenerator
	maxRetries int
}

func NewShortenerService(store *store.MemoryStore, generator *generator.RandomGenerator) *ShortenerService {
	return &ShortenerService{
		store:      store,
		generator:  generator,
		maxRetries: 3,
	}
}

func (s *ShortenerService) Create(ctx context.Context, url string) (string, error) {
	if existingCode, exists := s.store.GetByURL(ctx, url); exists {
		return existingCode, nil
	}

	for i := 0; i < s.maxRetries; i++ {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		code := s.generator.Generate()

		err := s.store.Save(ctx, code, url)
		if err == nil {
			return code, nil
		}

		if !errors.Is(err, models.ErrCollision) {
			return "", err
		}
	}

	return "", models.ErrFailedToGenerate
}

func (s *ShortenerService) Resolve(ctx context.Context, code string) (string, bool) {
	select {
	case <-ctx.Done():
		return "", false
	default:
	}

	return s.store.Load(ctx, code)
}
