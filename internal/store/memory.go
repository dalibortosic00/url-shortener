package store

import (
	"context"
	"sync"

	"github.com/dalibortosic00/url-shortener/internal/models"
)

type MemoryStore struct {
	mux       sync.RWMutex
	codeToUrl map[string]string
	urlToCode map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		codeToUrl: make(map[string]string),
		urlToCode: make(map[string]string),
	}
}

func (s *MemoryStore) Save(ctx context.Context, code string, url string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, exists := s.codeToUrl[code]; exists {
		return models.ErrCollision
	}

	s.codeToUrl[code] = url
	s.urlToCode[url] = code
	return nil
}

func (s *MemoryStore) Load(ctx context.Context, code string) (string, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	url, ok := s.codeToUrl[code]
	return url, ok
}

func (s *MemoryStore) GetByURL(ctx context.Context, url string) (string, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	code, ok := s.urlToCode[url]
	return code, ok
}
