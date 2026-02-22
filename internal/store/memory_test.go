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

	t.Run("Collision Error", func(t *testing.T) {
		code := "collision"
		s.SaveLink(ctx, &models.LinkRecord{
			Code:      code,
			URL:       "https://example.com",
			CreatedAt: time.Now(),
		})

		err := s.SaveLink(ctx, &models.LinkRecord{
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
			_ = s.SaveLink(ctx, &models.LinkRecord{
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
			_, _ = s.LoadLink(ctx, code)
		}(i)
	}

	wg.Wait()
}
