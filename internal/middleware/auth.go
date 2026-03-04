package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/dalibortosic00/url-shortener/internal/models"
	"github.com/dalibortosic00/url-shortener/internal/request"
	"github.com/dalibortosic00/url-shortener/internal/util"
)

func getAPIKey(header http.Header) (string, error) {
	authHeader := header.Get("Authorization")

	// If no header exists, it's a guest. No error needed.
	if authHeader == "" {
		return "", nil
	}

	// If a header exists, it must follow the "Bearer <key>" format.
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("invalid Authorization header format")
	}

	return parts[1], nil
}

type UserProvider interface {
	GetByAPIKey(ctx context.Context, apiKey string) (*models.User, error)
}

type AuthMiddleware struct {
	provider UserProvider
}

func NewAuthMiddleware(provider UserProvider) *AuthMiddleware {
	return &AuthMiddleware{provider: provider}
}

func (am *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth, err := getAPIKey(r.Header)
		if err != nil {
			util.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		if auth == "" {
			next.ServeHTTP(w, r)
			return
		}

		user, err := am.provider.GetByAPIKey(r.Context(), auth)
		if err != nil {
			if errors.Is(err, models.ErrRecordNotFound) {
				util.RespondWithError(w, http.StatusUnauthorized, "Invalid API key")
				return
			}
			util.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		ctx := request.WithUserID(r.Context(), user.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
