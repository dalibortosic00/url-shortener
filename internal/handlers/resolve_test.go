package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestResolveHandler_Resolve(t *testing.T) {
	tests := []struct {
		name             string
		code             string
		setup            func(svc *MockLinkResolver)
		expectedStatus   int
		expectedLocation string
	}{
		{
			name: "Successful Redirect",
			code: "abc123",
			setup: func(svc *MockLinkResolver) {
				svc.EXPECT().Resolve(mock.Anything, "abc123").Return("https://google.com", true)
			},
			expectedStatus:   http.StatusFound,
			expectedLocation: "https://google.com",
		},
		{
			name: "Code Not Found",
			code: "nonexistent",
			setup: func(svc *MockLinkResolver) {
				svc.EXPECT().Resolve(mock.Anything, "nonexistent").Return("", false)
			},
			expectedStatus:   http.StatusNotFound,
			expectedLocation: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMockLinkResolver(t)
			tt.setup(svc)

			h := NewResolveHandler(svc)

			req := httptest.NewRequest(http.MethodGet, "/"+tt.code, nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("code", tt.code)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			h.Resolve(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.expectedStatus, res.StatusCode)

			if tt.expectedLocation != "" {
				assert.Equal(t, tt.expectedLocation, res.Header.Get("Location"))
			}
		})
	}
}
