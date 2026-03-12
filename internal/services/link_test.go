package services

import (
	"context"
	"errors"
	"testing"

	"github.com/dalibortosic00/url-shortener/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLinkService_Create(t *testing.T) {
	testUrl := "http://example.com"
	errUnexpected := errors.New("unexpected error")

	tests := []struct {
		name         string
		url          string
		ownerID      string
		setup        func(p *MockLinkStore, r *MockLinkStore, g *MockCodeGenerator)
		expectedCode string
		expectedErr  error
	}{
		{
			name: "Successful - Public (New URL)",
			setup: func(p *MockLinkStore, _ *MockLinkStore, g *MockCodeGenerator) {
				p.EXPECT().GetCodeByURL(mock.Anything, testUrl).Return("", false)
				g.EXPECT().Generate(6).Return("abc123", nil)
				p.EXPECT().SaveLink(mock.Anything, mock.AnythingOfType("*models.LinkRecord")).Return(nil)
			},
			expectedCode: "abc123",
		},
		{
			name:    "Successful - Private",
			ownerID: "user123",
			setup: func(_ *MockLinkStore, r *MockLinkStore, g *MockCodeGenerator) {
				g.EXPECT().Generate(6).Return("def456", nil)
				r.EXPECT().SaveLink(mock.Anything, mock.AnythingOfType("*models.LinkRecord")).Return(nil)
			},
			expectedCode: "def456",
		},
		{
			name: "Public URL Exists (Deduplication)",
			setup: func(p *MockLinkStore, _ *MockLinkStore, _ *MockCodeGenerator) {
				p.EXPECT().GetCodeByURL(mock.Anything, testUrl).Return("existing123", true)
			},
			expectedCode: "existing123",
		},
		{
			name: "Public Collision (Exhaust Retries)",
			setup: func(p *MockLinkStore, _ *MockLinkStore, g *MockCodeGenerator) {
				p.EXPECT().GetCodeByURL(mock.Anything, testUrl).Return("", false)
				g.EXPECT().Generate(6).Return("collision", nil).Times(3)
				p.EXPECT().SaveLink(mock.Anything, mock.AnythingOfType("*models.LinkRecord")).Return(models.ErrCollision).Times(3)
			},
			expectedErr: models.ErrFailedToGenerate,
		},
		{
			name: "Successful - Recovery after Collision",
			setup: func(p *MockLinkStore, _ *MockLinkStore, g *MockCodeGenerator) {
				p.EXPECT().GetCodeByURL(mock.Anything, testUrl).Return("", false)

				g.EXPECT().Generate(6).Return("code1", nil).Once()
				p.EXPECT().SaveLink(mock.Anything, mock.AnythingOfType("*models.LinkRecord")).Return(models.ErrCollision).Once()

				g.EXPECT().Generate(6).Return("code2", nil).Once()
				p.EXPECT().SaveLink(mock.Anything, mock.AnythingOfType("*models.LinkRecord")).Return(nil).Once()
			},
			expectedCode: "code2",
		},
		{
			name: "Generator Error",
			setup: func(p *MockLinkStore, _ *MockLinkStore, g *MockCodeGenerator) {
				p.EXPECT().GetCodeByURL(mock.Anything, testUrl).Return("", false)
				g.EXPECT().Generate(6).Return("", errUnexpected)
			},
			expectedErr: errUnexpected,
		},
		{
			name: "Unexpected Store Error",
			setup: func(p *MockLinkStore, _ *MockLinkStore, g *MockCodeGenerator) {
				p.EXPECT().GetCodeByURL(mock.Anything, testUrl).Return("", false)
				g.EXPECT().Generate(6).Return("abc123", nil)
				p.EXPECT().SaveLink(mock.Anything, mock.AnythingOfType("*models.LinkRecord")).Return(errUnexpected)
			},
			expectedErr: errUnexpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewMockLinkStore(t)
			rs := NewMockLinkStore(t)
			cg := NewMockCodeGenerator(t)

			if tt.setup != nil {
				tt.setup(ps, rs, cg)
			}

			svc := NewLinkService(ps, rs, cg)
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

func TestLinkService_Resolve(t *testing.T) {
	tests := []struct {
		name          string
		code          string
		setup         func(p *MockLinkStore, r *MockLinkStore)
		expectedURL   string
		expectedFound bool
	}{
		{
			name: "Found in Public Store",
			code: "public123",
			setup: func(p *MockLinkStore, _ *MockLinkStore) {
				p.EXPECT().LoadLink(mock.Anything, "public123").Return(&models.LinkRecord{URL: "http://public.com"}, true)
			},
			expectedURL:   "http://public.com",
			expectedFound: true,
		},
		{
			name: "Found in Private Store",
			code: "private123",
			setup: func(p *MockLinkStore, r *MockLinkStore) {
				p.EXPECT().LoadLink(mock.Anything, "private123").Return(nil, false)
				r.EXPECT().LoadLink(mock.Anything, "private123").Return(&models.LinkRecord{URL: "http://private.com"}, true)
			},
			expectedURL:   "http://private.com",
			expectedFound: true,
		},
		{
			name: "Not Found in Any Store",
			code: "missing123",
			setup: func(p *MockLinkStore, r *MockLinkStore) {
				p.EXPECT().LoadLink(mock.Anything, "missing123").Return(nil, false)
				r.EXPECT().LoadLink(mock.Anything, "missing123").Return(nil, false)
			},
			expectedURL:   "",
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewMockLinkStore(t)
			rs := NewMockLinkStore(t)

			if tt.setup != nil {
				tt.setup(ps, rs)
			}

			svc := NewLinkService(ps, rs, nil)
			url, found := svc.Resolve(context.Background(), tt.code)

			assert.Equal(t, tt.expectedURL, url)
			assert.Equal(t, tt.expectedFound, found)
		})
	}
}
