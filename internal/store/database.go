package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/dalibortosic00/url-shortener/internal/models"
	"github.com/lib/pq"
)

// Compile-time check that DatabaseStore implements Store interface
var _ Store = (*DatabaseStore)(nil)

type DatabaseStore struct {
	db *sql.DB
}

func NewDatabaseStore(db *sql.DB) *DatabaseStore {
	return &DatabaseStore{db: db}
}

func (s *DatabaseStore) Save(ctx context.Context, record *models.LinkRecord) error {
	query := `
		INSERT INTO links (code, url, created_at)
		VALUES ($1, $2, $3)
	`

	_, err := s.db.ExecContext(ctx, query, record.Code, record.URL, record.CreatedAt)
	if err != nil {
		// Check for unique constraint violation on code
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return models.ErrCollision
		}
		return err
	}

	return nil
}

func (s *DatabaseStore) Load(ctx context.Context, code string) (string, bool) {
	query := `SELECT url FROM links WHERE code = $1`

	var url string
	err := s.db.QueryRowContext(ctx, query, code).Scan(&url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false
		}
		return "", false
	}

	return url, true
}

func (s *DatabaseStore) GetByURL(ctx context.Context, url string) (string, bool) {
	// Database store does NOT deduplicate - always return false
	// This allows authorized users to create multiple codes for the same URL
	return "", false
}
