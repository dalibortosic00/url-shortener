package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/dalibortosic00/url-shortener/internal/models"
	"github.com/lib/pq"
)

type DatabaseStore struct {
	db *sql.DB
}

func NewDatabaseStore(db *sql.DB) *DatabaseStore {
	return &DatabaseStore{db: db}
}

func (s *DatabaseStore) SaveLink(ctx context.Context, record *models.LinkRecord) error {
	query := `
		INSERT INTO links (code, url, owner_id, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := s.db.ExecContext(ctx, query, record.Code, record.URL, record.OwnerID, record.CreatedAt)
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

func (s *DatabaseStore) LoadLink(ctx context.Context, code string) (*models.LinkRecord, bool) {
	query := `SELECT code, url, owner_id, created_at FROM links WHERE code = $1`

	var record models.LinkRecord
	err := s.db.QueryRowContext(ctx, query, code).Scan(&record.Code, &record.URL, &record.OwnerID, &record.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false
		}
		return nil, false
	}

	return &record, true
}

func (s *DatabaseStore) GetLinksByOwner(ctx context.Context, ownerID string) ([]models.LinkRecord, error) {
	query := `SELECT code, url, owner_id, created_at FROM links WHERE owner_id = $1`

	rows, err := s.db.QueryContext(ctx, query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := make([]models.LinkRecord, 0)
	for rows.Next() {
		var record models.LinkRecord
		if err := rows.Scan(&record.Code, &record.URL, &record.OwnerID, &record.CreatedAt); err != nil {
			return nil, err
		}
		links = append(links, record)
	}

	return links, nil
}

func (s *DatabaseStore) GetCodeByURL(ctx context.Context, url string) (string, bool) {
	query := `SELECT code FROM links WHERE url = $1 AND owner_id IS NULL LIMIT 1`

	var code string
	err := s.db.QueryRowContext(ctx, query, url).Scan(&code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false
		}
		return "", false
	}

	return code, true
}

func (s *DatabaseStore) SaveUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, name, api_key)
		VALUES ($1, $2, $3)
	`

	_, err := s.db.ExecContext(ctx, query, user.ID, user.Name, user.APIKey)
	return err
}

func (s *DatabaseStore) GetUserByAPIKey(ctx context.Context, apiKey string) (*models.User, error) {
	query := `SELECT id, name FROM users WHERE api_key = $1`

	var user models.User
	err := s.db.QueryRowContext(ctx, query, apiKey).Scan(&user.ID, &user.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrRecordNotFound
		}
		return nil, err
	}

	user.APIKey = apiKey
	return &user, nil
}
