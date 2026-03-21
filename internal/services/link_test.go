package services

import (
	"context"
	"errors"
	"testing"

	"github.com/dalibortosic00/url-shortener/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func ptr[T any](v T) *T {
	return &v
}

func TestLinkService_Create(t *testing.T) {
	testUrl := "http://example.com"
	errUnexpected := errors.New("unexpected error")

	tests := []struct {
		name         string
		url          string
		ownerID      *string
		setup        func(s *MockLinkStore, g *MockCodeGenerator)
		expectedCode string
		expectedErr  error
	}{
		{
			name:    "Successful",
			ownerID: ptr("user123"),
			setup: func(s *MockLinkStore, g *MockCodeGenerator) {
				g.EXPECT().Generate(6).Return("def456", nil)
				s.EXPECT().SaveLink(mock.Anything, mock.AnythingOfType("*models.LinkRecord")).Return(nil)
			},
			expectedCode: "def456",
		},
		{
			name:    "Public URL Exists (Deduplication)",
			ownerID: nil,
			setup: func(s *MockLinkStore, _ *MockCodeGenerator) {
				s.EXPECT().GetCodeByURL(mock.Anything, testUrl).Return("existing123", true)
			},
			expectedCode: "existing123",
		},
		{
			name:    "Public Collision (Exhaust Retries)",
			ownerID: nil,
			setup: func(s *MockLinkStore, g *MockCodeGenerator) {
				s.EXPECT().GetCodeByURL(mock.Anything, testUrl).Return("", false)
				g.EXPECT().Generate(6).Return("collision", nil).Times(3)
				s.EXPECT().SaveLink(mock.Anything, mock.AnythingOfType("*models.LinkRecord")).Return(models.ErrCollision).Times(3)
			},
			expectedErr: models.ErrFailedToGenerate,
		},
		{
			name:    "Successful - Recovery after Collision",
			ownerID: nil,
			setup: func(s *MockLinkStore, g *MockCodeGenerator) {
				s.EXPECT().GetCodeByURL(mock.Anything, testUrl).Return("", false)

				g.EXPECT().Generate(6).Return("code1", nil).Once()
				s.EXPECT().SaveLink(mock.Anything, mock.AnythingOfType("*models.LinkRecord")).Return(models.ErrCollision).Once()

				g.EXPECT().Generate(6).Return("code2", nil).Once()
				s.EXPECT().SaveLink(mock.Anything, mock.AnythingOfType("*models.LinkRecord")).Return(nil).Once()
			},
			expectedCode: "code2",
		},
		{
			name: "Generator Error",
			setup: func(s *MockLinkStore, g *MockCodeGenerator) {
				s.EXPECT().GetCodeByURL(mock.Anything, testUrl).Return("", false)
				g.EXPECT().Generate(6).Return("", errUnexpected)
			},
			expectedErr: errUnexpected,
		},
		{
			name: "Unexpected Store Error",
			setup: func(s *MockLinkStore, g *MockCodeGenerator) {
				s.EXPECT().GetCodeByURL(mock.Anything, testUrl).Return("", false)
				g.EXPECT().Generate(6).Return("abc123", nil)
				s.EXPECT().SaveLink(mock.Anything, mock.AnythingOfType("*models.LinkRecord")).Return(errUnexpected)
			},
			expectedErr: errUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := NewMockLinkStore(t)
			cg := NewMockCodeGenerator(t)

			if tt.setup != nil {
				tt.setup(ls, cg)
			}

			svc := NewLinkService(ls, cg)
			code, err := svc.Create(context.Background(), testUrl, tt.ownerID)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedCode, code)
		})
	}
}

func TestLinkService_CreateCustom(t *testing.T) {
	testUrl := "http://example.com"
	customCode := "mylink"
	errUnexpected := errors.New("unexpected error")

	tests := []struct {
		name        string
		url         string
		customCode  string
		ownerID     *string
		setup       func(s *MockLinkStore)
		expectedErr error
	}{
		{
			name:       "Successful Custom Code Creation",
			url:        testUrl,
			customCode: customCode,
			ownerID:    ptr("user123"),
			setup: func(s *MockLinkStore) {
				s.EXPECT().SaveLink(mock.Anything, mock.AnythingOfType("*models.LinkRecord")).Return(nil)
			},
		},
		{
			name: "Store Error",
			setup: func(s *MockLinkStore) {
				s.EXPECT().SaveLink(mock.Anything, mock.AnythingOfType("*models.LinkRecord")).Return(errUnexpected)
			},
			expectedErr: errUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := NewMockLinkStore(t)

			if tt.setup != nil {
				tt.setup(ls)
			}

			svc := NewLinkService(ls, nil)
			code, err := svc.CreateCustom(context.Background(), tt.url, tt.customCode, tt.ownerID)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.customCode, code)
		})
	}
}

func TestLinkService_Resolve(t *testing.T) {
	tests := []struct {
		name          string
		code          string
		setup         func(s *MockLinkStore)
		expectedURL   string
		expectedFound bool
	}{
		{
			name: "Found",
			code: "private123",
			setup: func(s *MockLinkStore) {
				s.EXPECT().LoadLink(mock.Anything, "private123").Return(&models.LinkRecord{URL: "http://private.com"}, true)
			},
			expectedURL:   "http://private.com",
			expectedFound: true,
		},
		{
			name: "Not Found",
			code: "missing123",
			setup: func(s *MockLinkStore) {
				s.EXPECT().LoadLink(mock.Anything, "missing123").Return(nil, false)
			},
			expectedURL:   "",
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := NewMockLinkStore(t)

			if tt.setup != nil {
				tt.setup(ls)
			}

			svc := NewLinkService(ls, nil)
			url, found := svc.Resolve(context.Background(), tt.code)

			assert.Equal(t, tt.expectedURL, url)
			assert.Equal(t, tt.expectedFound, found)
		})
	}
}
