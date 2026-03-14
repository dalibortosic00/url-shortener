package helpers

import (
	"database/sql"
	"testing"

	"github.com/dalibortosic00/url-shortener/internal/store"
)

func NewTestDB(t *testing.T, connStr string) *sql.DB {
	t.Helper()

	db, err := store.InitDB(connStr)
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	return db
}

func Reset(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`TRUNCATE TABLE links, users RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatalf("reset db: %v", err)
	}
}
