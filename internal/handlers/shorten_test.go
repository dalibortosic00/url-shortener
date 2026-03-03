package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dalibortosic00/url-shortener/internal/models"
	"github.com/dalibortosic00/url-shortener/internal/request"
)

type mockLinkCreator struct {
	createFunc func(ctx context.Context, url, ownerID string) (string, error)
}

func (m *mockLinkCreator) Create(ctx context.Context, url, ownerID string) (string, error) {
	return m.createFunc(ctx, url, ownerID)
}

func TestShortenRequestValidate(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		forbiddenHost string
		wantErr       bool
		wantMsg       string
	}{
		{"Valid URL", "google.com", "localhost", false, ""},
		{"Empty URL", "", "localhost", true, errEmptyURL},
		{"Invalid Format", "google.com:invalid-port", "localhost", true, errInvalidFormat},
		{"Missing Host", "https:///path", "localhost", true, errInvalidDomain},
		{"Missing TLD", "http://mysite", "localhost", true, errInvalidTLD},
		{"Forbidden Host", "http://localhost/test", "localhost", true, errForbiddenDomain},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &shortenRequest{URL: tt.input}
			msg, ok := req.validate(tt.forbiddenHost)

			t.Logf("Input: %s | Resulting URL: %s | Msg: %s", tt.input, req.URL, msg)

			wantOK := !tt.wantErr

			if ok != wantOK {
				t.Fatalf("ok = %v; want %v", ok, wantOK)
			}

			if tt.wantErr && msg != tt.wantMsg {
				t.Errorf("msg = %q; want %q", msg, tt.wantMsg)
			}
		})
	}
}

func TestShortenHandler_Shorten(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		serviceMock    func() *mockLinkCreator
		ctxModifier    func(context.Context) context.Context
		expectedStatus int
		expectedInBody string
	}{
		{
			name:           "Success",
			requestBody:    `{"url": "https://google.com"}`,
			expectedStatus: http.StatusOK,
			serviceMock: func() *mockLinkCreator {
				return &mockLinkCreator{
					createFunc: func(ctx context.Context, url, ownerID string) (string, error) {
						return "abc123", nil
					},
				}
			},
			expectedInBody: "short_url",
		},
		{
			name:        "Success with OwnerID",
			requestBody: `{"url": "https://google.com"}`,
			serviceMock: func() *mockLinkCreator {
				return &mockLinkCreator{
					createFunc: func(ctx context.Context, url, ownerID string) (string, error) {
						if ownerID != "user-123" {
							return "", errors.New("expected ownerID user-123")
						}
						return "abc123", nil
					},
				}
			},
			ctxModifier: func(ctx context.Context) context.Context {
				return request.WithUserID(ctx, "user-123")
			},
			expectedStatus: http.StatusOK,
			expectedInBody: "short_url",
		},
		{
			name:           "Invalid JSON",
			requestBody:    `{"url": "https://google.com"`,
			expectedStatus: http.StatusBadRequest,
			serviceMock: func() *mockLinkCreator {
				return &mockLinkCreator{}
			},
			expectedInBody: "Invalid JSON",
		},
		{
			name:           "Validation Failure",
			requestBody:    `{"url": "not-a-url"}`,
			expectedStatus: http.StatusBadRequest,
			serviceMock: func() *mockLinkCreator {
				return &mockLinkCreator{}
			},
			expectedInBody: "valid TLD",
		},
		{
			name:           "Service Error",
			requestBody:    `{"url": "https://google.com"}`,
			expectedStatus: http.StatusInternalServerError,
			serviceMock: func() *mockLinkCreator {
				return &mockLinkCreator{
					createFunc: func(ctx context.Context, url, ownerID string) (string, error) {
						return "", models.ErrFailedToGenerate
					},
				}
			},
			expectedInBody: "Internal service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.serviceMock()
			h := NewShortenHandler(svc, "https://sho.rt")

			req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			if tt.ctxModifier != nil {
				req = req.WithContext(tt.ctxModifier(req.Context()))
			}

			w := httptest.NewRecorder()

			h.Shorten(w, req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d; got %d", tt.expectedStatus, res.StatusCode)
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
