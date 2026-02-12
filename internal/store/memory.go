package store

import (
	"context"
	"sync"

	"github.com/dalibortosic00/url-shortener/internal/models"
)

type MemoryStore struct {
	mux       sync.RWMutex
	records   map[string]*models.LinkRecord
	urlToCode map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		records:   make(map[string]*models.LinkRecord),
		urlToCode: make(map[string]string),
	}
}

func (s *MemoryStore) Save(ctx context.Context, record *models.LinkRecord) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, exists := s.records[record.Code]; exists {
		return models.ErrCollision
	}

	s.records[record.Code] = record
	s.urlToCode[record.URL] = record.Code
	return nil
}

func (s *MemoryStore) Load(ctx context.Context, code string) (string, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	record, ok := s.records[code]
	if !ok {
		return "", false
	}
	return record.URL, true
}

func (s *MemoryStore) GetByURL(ctx context.Context, url string) (string, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	code, ok := s.urlToCode[url]
	return code, ok
}
