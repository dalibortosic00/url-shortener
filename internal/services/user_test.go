package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserService_Create(t *testing.T) {
	testName := "Test User"
	errUnexpected := errors.New("unexpected error")

	tests := []struct {
		name        string
		setup       func(s *MockUserStore, g *MockCodeGenerator)
		expectedKey string
		expectedErr error
	}{
		{
			name: "Successful",
			setup: func(s *MockUserStore, g *MockCodeGenerator) {
				g.EXPECT().Generate(32).Return("api-key", nil)
				s.EXPECT().SaveUser(mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
			},
			expectedKey: "api-key",
		},
		{
			name: "Code Generation Failure",
			setup: func(s *MockUserStore, g *MockCodeGenerator) {
				g.EXPECT().Generate(32).Return("", errUnexpected)
			},
			expectedErr: errUnexpected,
		},
		{
			name: "Unexpected Store Error",
			setup: func(s *MockUserStore, g *MockCodeGenerator) {
				g.EXPECT().Generate(32).Return("generated-api-key", nil)
				s.EXPECT().SaveUser(mock.Anything, mock.AnythingOfType("*models.User")).Return(errUnexpected)
			},
			expectedErr: errUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockUserStore)
			mockGenerator := new(MockCodeGenerator)
			tt.setup(mockStore, mockGenerator)

			service := NewUserService(mockStore, mockGenerator)
			apiKey, err := service.Create(context.Background(), testName)

			assert.Equal(t, tt.expectedKey, apiKey)
			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}
