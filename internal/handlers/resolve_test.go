package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

type mockLinkResolver struct {
	resolveFunc func(ctx context.Context, code string) (string, bool)
}

func (m *mockLinkResolver) Resolve(ctx context.Context, code string) (string, bool) {
	return m.resolveFunc(ctx, code)
}

func TestResolveHandler_Resolve(t *testing.T) {
	tests := []struct {
		name             string
		code             string
		serviceMock      func() *mockLinkResolver
		expectedStatus   int
		expectedLocation string
	}{
		{
			name: "Successful Redirect",
			code: "abc123",
			serviceMock: func() *mockLinkResolver {
				return &mockLinkResolver{
					resolveFunc: func(ctx context.Context, code string) (string, bool) {
						if code == "" {
							return "", false
						}
						return "https://google.com", true
					}}
			},
			expectedStatus:   http.StatusFound,
			expectedLocation: "https://google.com",
		},
		{
			name: "Code Not Found",
			code: "nonexistent",
			serviceMock: func() *mockLinkResolver {
				return &mockLinkResolver{
					resolveFunc: func(ctx context.Context, code string) (string, bool) {
						return "", false
					}}
			},
			expectedStatus:   http.StatusNotFound,
			expectedLocation: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.serviceMock()
			h := NewResolveHandler(svc)

			req := httptest.NewRequest(http.MethodGet, "/"+tt.code, nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("code", tt.code)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			h.Resolve(w, req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.StatusCode)
			}

			if tt.expectedLocation != "" {
				location := res.Header.Get("Location")
				if location != tt.expectedLocation {
					t.Errorf("expected Location header %s, got %s", tt.expectedLocation, location)
				}
			}
		})
	}
}
