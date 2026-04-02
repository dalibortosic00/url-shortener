package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dalibortosic00/url-shortener/internal/models"
	"github.com/dalibortosic00/url-shortener/internal/request"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeleteHandler_Delete(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		userID         string
		setup          func(svc *MockLinkDeleter)
		expectedStatus int
		expectedInBody []string
	}{
		{
			name:   "Successful Delete",
			code:   "abc123",
			userID: "user123",
			setup: func(svc *MockLinkDeleter) {
				svc.EXPECT().Delete(mock.Anything, "abc123", "user123").Return(nil)
			},
			expectedStatus: http.StatusNoContent,
			expectedInBody: []string{},
		},
		{
			name:           "Unauthorized - No User ID",
			code:           "abc123",
			userID:         "",
			setup:          func(svc *MockLinkDeleter) {},
			expectedStatus: http.StatusUnauthorized,
			expectedInBody: []string{"Authentication required"},
		},
		{
			name:   "Not Found",
			code:   "nonexistent",
			userID: "user123",
			setup: func(svc *MockLinkDeleter) {
				svc.EXPECT().Delete(mock.Anything, "nonexistent", "user123").Return(models.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedInBody: []string{"Short code not found"},
		},
		{
			name:   "Service Error",
			code:   "abc123",
			userID: "user123",
			setup: func(svc *MockLinkDeleter) {
				svc.EXPECT().Delete(mock.Anything, "abc123", "user123").Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedInBody: []string{"Internal service error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := NewMockLinkDeleter(t)
			tt.setup(mockService)

			handler := NewDeleteHandler(mockService)

			req := httptest.NewRequest("DELETE", "/"+tt.code, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("code", tt.code)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			if tt.userID != "" {
				req = req.WithContext(request.WithUserID(req.Context(), tt.userID))
			}

			rec := httptest.NewRecorder()
			handler.Delete(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			for _, expected := range tt.expectedInBody {
				assert.Contains(t, rec.Body.String(), expected)
			}
		})
	}
}
