package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockUserCreator struct {
	createFunc func(ctx context.Context, name string) (string, error)
}

func (m *mockUserCreator) Create(ctx context.Context, name string) (string, error) {
	return m.createFunc(ctx, name)
}

func TestRegisterHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		serviceMock    func() *mockUserCreator
		expectedStatus int
		expectedInBody string
	}{
		{
			name:        "Success",
			requestBody: `{"name": "Alice"}`,
			serviceMock: func() *mockUserCreator {
				return &mockUserCreator{
					createFunc: func(ctx context.Context, name string) (string, error) {
						return "user-123", nil
					},
				}
			},
			expectedStatus: http.StatusCreated,
			expectedInBody: "registered successfully",
		},
		{
			name:        "Invalid JSON",
			requestBody: `{"name": "Alice"`,
			serviceMock: func() *mockUserCreator {
				return &mockUserCreator{}
			},
			expectedStatus: http.StatusBadRequest,
			expectedInBody: "Invalid JSON payload",
		},
		{
			name:        "Service Error",
			requestBody: `{"name": "Alice"}`,
			serviceMock: func() *mockUserCreator {
				return &mockUserCreator{
					createFunc: func(ctx context.Context, name string) (string, error) {
						return "", errors.New("error creating user")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedInBody: "Internal service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.serviceMock()
			h := NewRegisterHandler(svc)

			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			h.Register(w, req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.StatusCode)
			}

			bodyBytes, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("could not read response body: %v", err)
			}
			body := string(bodyBytes)

			if tt.expectedInBody != "" && !strings.Contains(body, tt.expectedInBody) {
				t.Errorf("expected response to contain %q; got %q", tt.expectedInBody, body)
			}
		})
	}
}
