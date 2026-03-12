package store

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/dalibortosic00/url-shortener/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStore(t *testing.T) {
	t.Run("Save and Load", func(t *testing.T) {
		s := NewMemoryStore()
		ctx := context.Background()
		code, url := "abc123", "https://example.com"

		err := s.SaveLink(ctx, &models.LinkRecord{
			Code:      code,
			URL:       url,
			CreatedAt: time.Now(),
		})
		require.NoError(t, err)

		record, ok := s.LoadLink(ctx, code)
		require.True(t, ok)
		assert.Equal(t, url, record.URL)

		loadedCode, ok := s.GetCodeByURL(ctx, url)
		require.True(t, ok)
		assert.Equal(t, code, loadedCode)
	})

	t.Run("Load Non-Existent Code", func(t *testing.T) {
		s := NewMemoryStore()
		ctx := context.Background()

		_, ok := s.LoadLink(ctx, "nonexistent")
		assert.False(t, ok)
	})

	t.Run("GetCodeByURL Non-Existent URL", func(t *testing.T) {
		s := NewMemoryStore()
		ctx := context.Background()

		_, ok := s.GetCodeByURL(ctx, "https://nonexistent.com")
		assert.False(t, ok)
	})

	t.Run("Collision Error", func(t *testing.T) {
		s := NewMemoryStore()
		ctx := context.Background()
		code := "collision"

		err := s.SaveLink(ctx, &models.LinkRecord{
			Code:      code,
			URL:       "https://example.com",
			CreatedAt: time.Now(),
		})
		require.NoError(t, err)

		err = s.SaveLink(ctx, &models.LinkRecord{
			Code:      code,
			URL:       "https://another.com",
			CreatedAt: time.Now(),
		})
		assert.ErrorIs(t, err, models.ErrCollision)
	})
}

func TestMemoryStore_Concurrent(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()
	var wg sync.WaitGroup

	for i := range 1000 {
		wg.Add(3)

		go func(n int) {
			defer wg.Done()
			_ = s.SaveLink(ctx, &models.LinkRecord{
				Code:      fmt.Sprintf("code-%d", n),
				URL:       fmt.Sprintf("url-%d", n),
				CreatedAt: time.Now(),
			})
		}(i)

		go func(n int) {
			defer wg.Done()
			_, _ = s.LoadLink(ctx, fmt.Sprintf("code-%d", n))
		}(i)

		go func(n int) {
			defer wg.Done()
			_, _ = s.GetCodeByURL(ctx, fmt.Sprintf("url-%d", n))
		}(i)
	}

	wg.Wait()
}
