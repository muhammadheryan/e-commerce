package transport

import (
	"context"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/muhammadheryan/e-commerce/application/user"
	"github.com/muhammadheryan/e-commerce/constant"
	"github.com/muhammadheryan/e-commerce/utils/errors"
)

// AuthMiddleware returns a middleware that validates JWT sessions using UserApp.
// It allows public endpoints (like /login, /register, /swagger/) without token.
func AuthMiddleware(userApp user.UserApp) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Public paths
			path := r.URL.Path
			if isPublicPath(path) {
				next.ServeHTTP(w, r)
				return
			}

			// Check Authorization header
			auth := r.Header.Get("Authorization")
			if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
				writeError(w, errors.SetCustomError(constant.ErrUnauthorize))
				return
			}
			token := strings.TrimPrefix(auth, "Bearer ")

			// Validate token via UserApp
			userID, err := userApp.ValidateToken(r.Context(), token)
			if err != nil {
				writeError(w, errors.SetCustomError(constant.ErrUnauthorize))
				return
			}

			// Embed userID into context
			ctx := context.WithValue(r.Context(), constant.UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// isPublicPath defines which endpoints are public (no auth required)
func isPublicPath(path string) bool {
	if strings.HasPrefix(path, "/swagger/") || strings.HasPrefix(path, "/internal/") {
		return true
	}
	if path == "/login" || path == "/register" {
		return true
	}

	return false
}
