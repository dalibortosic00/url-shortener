package store

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/dalibortosic00/url-shortener/internal/models"
)

func TestMemoryStore(t *testing.T) {
	t.Run("Save and Load", func(t *testing.T) {
		s := NewMemoryStore()
		ctx := context.Background()
		code, url := "abc123", "https://example.com"

		shortenedURL := &models.LinkRecord{
			Code:      code,
			URL:       url,
			CreatedAt: time.Now(),
		}

		if err := s.SaveLink(ctx, shortenedURL); err != nil {
			t.Fatalf("failed to save: %v", err)
		}

		record, ok := s.LoadLink(ctx, code)
		if !ok || record.URL != url {
			t.Fatalf("Load() = %v, %v; want %v, true", record, ok, url)
		}

		loadedCode, ok := s.GetCodeByURL(ctx, url)
		if !ok || loadedCode != code {
			t.Fatalf("GetByURL() = %v, %v; want %v, true", loadedCode, ok, code)
		}
	})

	t.Run("Load Non-Existent Code", func(t *testing.T) {
		s := NewMemoryStore()
		ctx := context.Background()

		if _, ok := s.LoadLink(ctx, "nonexistent"); ok {
			t.Fatal("expected Load() to return false for non-existent code")
		}
	})

	t.Run("GetCodeByURL Non-Existent URL", func(t *testing.T) {
		s := NewMemoryStore()
		ctx := context.Background()

		if _, ok := s.GetCodeByURL(ctx, "https://nonexistent.com"); ok {
			t.Fatal("expected GetCodeByURL() to return false for non-existent URL")
		}
	})

	t.Run("Collision Error", func(t *testing.T) {
		s := NewMemoryStore()
		ctx := context.Background()
		code := "collision"

		if err := s.SaveLink(ctx, &models.LinkRecord{
			Code:      code,
			URL:       "https://example.com",
			CreatedAt: time.Now(),
		}); err != nil {
			t.Fatalf("failed to save: %v", err)
		}

		if err := s.SaveLink(ctx, &models.LinkRecord{
			Code:      code,
			URL:       "https://another.com",
			CreatedAt: time.Now(),
		}); !errors.Is(err, models.ErrCollision) {
			t.Fatalf("expected ErrCollision, got %v", err)
		}
	})
}

func TestMemoryStore_Concurrent(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()
	var wg sync.WaitGroup

	iterations := 1000

	for i := range iterations {
		wg.Add(3)
		go func(n int) {
			defer wg.Done()
			code := fmt.Sprintf("code-%d", n)
			url := fmt.Sprintf("url-%d", n)
			_ = s.SaveLink(ctx, &models.LinkRecord{
				Code:      code,
				URL:       url,
				CreatedAt: time.Now(),
			})
		}(i)

		go func(n int) {
			defer wg.Done()
			code := fmt.Sprintf("code-%d", n)
			_, _ = s.LoadLink(ctx, code)
		}(i)

		go func(n int) {
			defer wg.Done()
			url := fmt.Sprintf("url-%d", n)
			_, _ = s.GetCodeByURL(ctx, url)
		}(i)
	}

	wg.Wait()
}
