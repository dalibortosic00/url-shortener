package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dalibortosic00/url-shortener/internal/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

			assert.Equal(t, !tt.wantErr, ok)

			if tt.wantErr {
				assert.Equal(t, tt.wantMsg, msg)
			}
		})
	}
}

func TestShortenHandler_Shorten(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		setup          func(svc *MockLinkCreator)
		ctxModifier    func(ctx context.Context) context.Context
		expectedStatus int
		expectedInBody string
	}{
		{
			name:        "Success",
			requestBody: `{"url": "https://google.com"}`,
			setup: func(svc *MockLinkCreator) {
				svc.EXPECT().Create(mock.Anything, "https://google.com", "").Return("abc123", nil)
			},
			expectedStatus: http.StatusOK,
			expectedInBody: "short_url",
		},
		{
			name:        "Success with OwnerID",
			requestBody: `{"url": "https://google.com"}`,
			setup: func(svc *MockLinkCreator) {
				svc.EXPECT().Create(mock.Anything, "https://google.com", "user-123").Return("abc123", nil)
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
			setup:          func(svc *MockLinkCreator) {},
			expectedStatus: http.StatusBadRequest,
			expectedInBody: "Invalid JSON",
		},
		{
			name:           "Validation Failure",
			requestBody:    `{"url": "not-a-url"}`,
			setup:          func(svc *MockLinkCreator) {},
			expectedStatus: http.StatusBadRequest,
			expectedInBody: "valid TLD",
		},
		{
			name:        "Service Error",
			requestBody: `{"url": "https://google.com"}`,
			setup: func(svc *MockLinkCreator) {
				svc.EXPECT().Create(mock.Anything, "https://google.com", "").Return("", errors.New("error creating short code"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedInBody: "Internal service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMockLinkCreator(t)
			tt.setup(svc)

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

			assert.Equal(t, tt.expectedStatus, res.StatusCode)

			bodyBytes, err := io.ReadAll(res.Body)
			assert.NoError(t, err)

			if tt.expectedInBody != "" {
				assert.Contains(t, string(bodyBytes), tt.expectedInBody)
			}
		})
	}
}
