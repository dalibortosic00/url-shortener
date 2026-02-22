package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/dalibortosic00/url-shortener/internal/store"
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

type AuthMiddleware struct {
	userStore store.UserStore
}

func NewAuthMiddleware(userStore store.UserStore) *AuthMiddleware {
	return &AuthMiddleware{userStore: userStore}
}

type contextKey string

const userContextKey = contextKey("user")

func UserIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(userContextKey).(string)
	return id
}

func (am *AuthMiddleware) Middleware(next http.HandlerFunc) http.HandlerFunc {
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

		user, err := am.userStore.GetUserByAPIKey(r.Context(), auth)
		if err != nil {
			if errors.Is(err, store.ErrRecordNotFound) {
				util.RespondWithError(w, http.StatusUnauthorized, "Invalid API key")
				return
			}
			util.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
