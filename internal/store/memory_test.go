package store

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/dalibortosic00/url-shortener/internal/models"
)

func TestMemoryStore(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	t.Run("Save and Load", func(t *testing.T) {
		code, url := "abc123", "https://example.com"

		shortenedURL := &models.LinkRecord{
			Code:      code,
			URL:       url,
			CreatedAt: time.Now(),
		}

		if err := s.Save(ctx, shortenedURL); err != nil {
			t.Fatalf("failed to save: %v", err)
		}

		loadedURL, ok := s.Load(ctx, code)
		if !ok || loadedURL != url {
			t.Fatalf("Load() = %v, %v; want %v, true", loadedURL, ok, url)
		}

		loadedCode, ok := s.GetByURL(ctx, url)
		if !ok || loadedCode != code {
			t.Fatalf("GetByURL() = %v, %v; want %v, true", loadedCode, ok, code)
		}
	})

	t.Run("Collision Error", func(t *testing.T) {
		code := "collision"
		s.Save(ctx, &models.LinkRecord{
			Code:      code,
			URL:       "https://example.com",
			CreatedAt: time.Now(),
		})

		err := s.Save(ctx, &models.LinkRecord{
			Code:      code,
			URL:       "https://another.com",
			CreatedAt: time.Now(),
		})
		if err != models.ErrCollision {
			t.Fatalf("expected ErrCollision, got %v", err)
		}
	})
}

func TestMemoryStore_Concurrent(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()
	wg := sync.WaitGroup{}

	iterations := 1000

	for i := range iterations {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			code := fmt.Sprintf("code-%d", n)
			url := fmt.Sprintf("url-%d", n)
			_ = s.Save(ctx, &models.LinkRecord{
				Code:      code,
				URL:       url,
				CreatedAt: time.Now(),
			})
		}(i)
	}

	for i := range iterations {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			code := fmt.Sprintf("code-%d", n)
			_, _ = s.Load(ctx, code)
		}(i)
	}

	wg.Wait()
}
