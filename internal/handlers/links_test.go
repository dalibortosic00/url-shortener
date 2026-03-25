package handlers

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dalibortosic00/url-shortener/internal/models"
	"github.com/dalibortosic00/url-shortener/internal/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const testBaseURL = "http://localhost:8080"

func TestLinksHandler_List(t *testing.T) {
	tests := []struct {
		name           string
		ownerID        string
		setup          func(svc *MockLinkLister)
		expectedStatus int
		expectedInBody []string
	}{
		{
			name:    "Successful List",
			ownerID: "user123",
			setup: func(svc *MockLinkLister) {
				svc.EXPECT().List(mock.Anything, "user123").Return([]models.LinkRecord{
					{Code: "abc123", URL: "https://google.com"},
					{Code: "def456", URL: "https://github.com"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedInBody: []string{
				testBaseURL + "/abc123",
				testBaseURL + "/def456",
			},
		},
		{
			name:    "Service Error",
			ownerID: "user123",
			setup: func(svc *MockLinkLister) {
				svc.EXPECT().List(mock.Anything, "user123").Return(nil, errors.New("error fetching links"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedInBody: []string{"Internal service error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMockLinkLister(t)
			tt.setup(svc)

			h := NewLinksHandler(svc, testBaseURL)

			req := httptest.NewRequest(http.MethodGet, "/links", nil)
			req = req.WithContext(request.WithUserID(req.Context(), tt.ownerID))
			w := httptest.NewRecorder()
			h.List(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.expectedStatus, res.StatusCode)
			bodyBytes, _ := io.ReadAll(res.Body)

			for _, expected := range tt.expectedInBody {
				assert.Contains(t, string(bodyBytes), expected)
			}
		})
	}
}
