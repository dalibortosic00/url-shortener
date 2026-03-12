package handlers

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		setup          func(svc *MockUserCreator)
		expectedStatus int
		expectedInBody string
	}{
		{
			name:        "Success",
			requestBody: `{"name": "Alice"}`,
			setup: func(svc *MockUserCreator) {
				svc.EXPECT().Create(mock.Anything, "Alice").Return("api-key", nil)
			},
			expectedStatus: http.StatusCreated,
			expectedInBody: "registered successfully",
		},
		{
			name:           "Invalid JSON",
			requestBody:    `{"name": "Alice"`,
			setup:          func(svc *MockUserCreator) {},
			expectedStatus: http.StatusBadRequest,
			expectedInBody: "Invalid JSON payload",
		},
		{
			name:        "Service Error",
			requestBody: `{"name": "Alice"}`,
			setup: func(svc *MockUserCreator) {
				svc.EXPECT().Create(mock.Anything, "Alice").Return("", errors.New("error creating user"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedInBody: "Internal service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMockUserCreator(t)
			tt.setup(svc)

			h := NewRegisterHandler(svc)

			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h.Register(w, req)

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
