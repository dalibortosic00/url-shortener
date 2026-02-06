package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dalibortosic00/url-shortener/internal/generator"
	"github.com/dalibortosic00/url-shortener/internal/services"
	"github.com/dalibortosic00/url-shortener/internal/store"
	"github.com/go-chi/chi/v5"
)

func TestResolveHandler_Resolve(t *testing.T) {
	ms := store.NewMemoryStore()
	gen := generator.NewRandomGenerator(6)
	svc := services.NewShortenerService(ms, gen)
	h := NewResolveHandler(svc)

	testCode := "abc123"
	testURL := "https://google.com"
	ms.Save(context.Background(), testCode, testURL)

	tests := []struct {
		name             string
		code             string
		expectedStatus   int
		expectedLocation string
	}{
		{
			name:             "Successful Redirect",
			code:             testCode,
			expectedStatus:   http.StatusFound,
			expectedLocation: testURL,
		},
		{
			name:             "Code Not Found",
			code:             "nonexistent",
			expectedStatus:   http.StatusNotFound,
			expectedLocation: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/"+tt.code, nil)

			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("code", tt.code)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			h.Resolve(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d; got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusFound {
				location := w.Header().Get("Location")
				if location != tt.expectedLocation {
					t.Errorf("expected redirect to %q; got %q", tt.expectedLocation, location)
				}
			}
		})
	}
}
