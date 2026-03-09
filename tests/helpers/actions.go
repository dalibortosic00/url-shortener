package helpers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type ShortenOption func(*http.Request)

func WithAPIKey(key string) ShortenOption {
	return func(req *http.Request) {
		req.Header.Set("Authorization", "Bearer "+key)
	}
}

func Shorten(t *testing.T, ts *httptest.Server, url string, opts ...ShortenOption) string {
	t.Helper()
	body := `{"url":"` + url + `"}`
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/shorten", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for _, opt := range opts {
		opt(req)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("shorten: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("shorten got %d, want 200", resp.StatusCode)
	}

	var result struct {
		ShortURL string `json:"short_url"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	parts := strings.Split(result.ShortURL, "/")
	return parts[len(parts)-1]
}

func Resolve(t *testing.T, ts *httptest.Server, code string) string {
	t.Helper()

	// don't follow the redirect automatically
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(ts.URL + "/" + code)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusMovedPermanently {
		t.Fatalf("resolve got %d, want redirect", resp.StatusCode)
	}

	return resp.Header.Get("Location")
}

func Register(t *testing.T, ts *httptest.Server, username string) string {
	t.Helper()
	body := `{"name":"` + username + `"}`
	resp, err := http.Post(ts.URL+"/register", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register got %d, want 201", resp.StatusCode)
	}

	var result struct {
		Message string `json:"message"`
		ApiKey  string `json:"api_key"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.ApiKey
}
