package helpers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type ShortenParams struct {
	URL        string
	CustomCode string
}

type ShortenOption func(*http.Request)

func WithAPIKey(key string) ShortenOption {
	return func(req *http.Request) {
		req.Header.Set("Authorization", "Bearer "+key)
	}
}

func Shorten(t *testing.T, ts *httptest.Server, params ShortenParams, opts ...ShortenOption) string {
	t.Helper()
	res := shortenRequest(t, ts, params, opts...)
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("shorten got %d, want 200", res.StatusCode)
	}

	var result struct {
		ShortURL string `json:"short_url"`
	}
	json.NewDecoder(res.Body).Decode(&result)
	parts := strings.Split(result.ShortURL, "/")
	return parts[len(parts)-1]
}

func ShortenRaw(t *testing.T, ts *httptest.Server, params ShortenParams, opts ...ShortenOption) *http.Response {
	t.Helper()
	return shortenRequest(t, ts, params, opts...)
}

func Resolve(t *testing.T, ts *httptest.Server, code string) string {
	t.Helper()
	res := resolveRequest(t, ts, code)
	defer res.Body.Close()

	if res.StatusCode != http.StatusFound {
		t.Fatalf("resolve got %d, want 302", res.StatusCode)
	}

	location, err := res.Location()
	if err != nil {
		t.Fatalf("resolve location: %v", err)
	}
	return location.String()
}

func ResolveRaw(t *testing.T, ts *httptest.Server, code string) *http.Response {
	t.Helper()
	return resolveRequest(t, ts, code)
}

func Register(t *testing.T, ts *httptest.Server, username string) string {
	t.Helper()
	res := registerRequest(t, ts, username)
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("register got %d, want 201", res.StatusCode)
	}

	var result struct {
		Message string `json:"message"`
		APIKey  string `json:"api_key"`
	}
	json.NewDecoder(res.Body).Decode(&result)
	return result.APIKey
}

func RegisterRaw(t *testing.T, ts *httptest.Server, username string) *http.Response {
	t.Helper()
	return registerRequest(t, ts, username)
}

func Links(t *testing.T, ts *httptest.Server, apiKey string) map[string]string {
	t.Helper()
	res := linksRequest(t, ts, apiKey)
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("links got %d, want 200", res.StatusCode)
	}

	var result map[string]string
	json.NewDecoder(res.Body).Decode(&result)
	return result
}

func LinksRaw(t *testing.T, ts *httptest.Server, apiKey string) *http.Response {
	t.Helper()
	return linksRequest(t, ts, apiKey)
}

func Delete(t *testing.T, ts *httptest.Server, code string, apiKey string) {
	t.Helper()
	res := deleteRequest(t, ts, code, apiKey)
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("delete got %d, want 204", res.StatusCode)
	}
}

func DeleteRaw(t *testing.T, ts *httptest.Server, code string, apiKey string) *http.Response {
	t.Helper()
	return deleteRequest(t, ts, code, apiKey)
}

func shortenRequest(t *testing.T, ts *httptest.Server, params ShortenParams, opts ...ShortenOption) *http.Response {
	t.Helper()
	body := `{"url":"` + params.URL + `", "custom_code":"` + params.CustomCode + `"}`
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/shorten", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for _, opt := range opts {
		opt(req)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("shorten request: %v", err)
	}
	return res
}

func resolveRequest(t *testing.T, ts *httptest.Server, code string) *http.Response {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/"+code, nil)
	// don't follow the redirect automatically
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	res, err := client.Do(req)
	if err != nil {
		t.Fatalf("resolve request: %v", err)
	}
	return res
}

func registerRequest(t *testing.T, ts *httptest.Server, username string) *http.Response {
	t.Helper()
	body := `{"name":"` + username + `"}`
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("register request: %v", err)
	}
	return res
}

func linksRequest(t *testing.T, ts *httptest.Server, apiKey string) *http.Response {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/links", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("links request: %v", err)
	}
	return res
}

func deleteRequest(t *testing.T, ts *httptest.Server, code string, apiKey string) *http.Response {
	t.Helper()
	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/"+code, nil)
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("delete request: %v", err)
	}
	return res
}
