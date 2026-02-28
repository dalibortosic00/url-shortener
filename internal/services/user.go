package services

import (
	"context"
	"fmt"

	"github.com/dalibortosic00/url-shortener/internal/generators"
	"github.com/dalibortosic00/url-shortener/internal/models"
	"github.com/dalibortosic00/url-shortener/internal/util"
)

// UserStore defines the interface for user storage, allowing retrieval of users by API key.
type UserStore interface {
	// Save persists a user record.
	SaveUser(ctx context.Context, user *models.User) error

	// GetUserByAPIKey retrieves a user by their API key.
	// Returns (user, nil) if found, (nil, ErrRecordNotFound) if not found, or (nil, error) on other errors.
	GetUserByAPIKey(ctx context.Context, apiKey string) (*models.User, error)
}

type UserService struct {
	userStore UserStore
	generator *generators.RandomGenerator
}

func NewUserService(userStore UserStore, generator *generators.RandomGenerator) *UserService {
	return &UserService{userStore: userStore,
		generator: generator,
	}
}

func (s *UserService) Create(ctx context.Context, name string) (string, error) {
	apiKey, err := s.generator.Generate(32)
	if err != nil {
		return "", fmt.Errorf("failed to generate api key: %w", err)
	}

	user := &models.User{
		ID:     util.GenerateUserID(),
		Name:   name,
		APIKey: apiKey,
	}

	if err := s.userStore.SaveUser(ctx, user); err != nil {
		return "", err
	}

	return apiKey, nil
}

func (s *UserService) GetByAPIKey(ctx context.Context, apiKey string) (*models.User, error) {
	return s.userStore.GetUserByAPIKey(ctx, apiKey)
}
