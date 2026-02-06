package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dalibortosic00/url-shortener/internal/generator"
	"github.com/dalibortosic00/url-shortener/internal/services"
	"github.com/dalibortosic00/url-shortener/internal/store"
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
	ms := store.NewMemoryStore()
	gen := generator.NewRandomGenerator(6)
	svc := services.NewShortenerService(ms, gen)
	h := NewShortenHandler(svc, "https://sho.rt")

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "Success",
			requestBody:    `{"url": "https://google.com"}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid JSON",
			requestBody:    `{"url": "https://google.com"`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Validation Failure",
			requestBody:    `{"url": "not-a-url"}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			h.Shorten(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d; got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var res map[string]string
				if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if _, ok := res["short_url"]; !ok {
					t.Error("response did not contain 'short_url'")
				}
			}
		})
	}
}
