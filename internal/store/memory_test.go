package store

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/dalibortosic00/url-shortener/internal/models"
)

func TestMemoryStore(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	t.Run("Save and Load", func(t *testing.T) {
		code, url := "abc123", "https://example.com"

		if err := s.Save(ctx, code, url); err != nil {
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
		s.Save(ctx, code, "https://example.com")

		err := s.Save(ctx, code, "https://another.com")
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
			_ = s.Save(ctx, code, url)
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
