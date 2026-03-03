package store

import (
	"context"
	"sync"

	"github.com/dalibortosic00/url-shortener/internal/models"
	"github.com/dalibortosic00/url-shortener/internal/services"
)

var _ services.LinkStore = (*MemoryStore)(nil)

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

func (s *MemoryStore) SaveLink(ctx context.Context, record *models.LinkRecord) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, exists := s.records[record.Code]; exists {
		return models.ErrCollision
	}

	s.records[record.Code] = record
	s.urlToCode[record.URL] = record.Code
	return nil
}

func (s *MemoryStore) LoadLink(ctx context.Context, code string) (*models.LinkRecord, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	record, ok := s.records[code]
	if !ok {
		return nil, false
	}
	return record, true
}

func (s *MemoryStore) GetCodeByURL(ctx context.Context, url string) (string, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	code, ok := s.urlToCode[url]
	return code, ok
}
