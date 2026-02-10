package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/atta/vulnpulse/pkg/auth"
	"github.com/atta/vulnpulse/pkg/logger"
)

type contextKey string

const (
	ContextKeyClaims contextKey = "claims"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware(authService *auth.Service, log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Warn("missing authorization header")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// Extract token from "Bearer <token>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				log.Warn("invalid authorization header format")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			token := parts[1]
			claims, err := authService.ValidateToken(token)
			if err != nil {
				log.Warn("invalid token", "error", err)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// Add claims to request context
			ctx := context.WithValue(r.Context(), ContextKeyClaims, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetClaims extracts JWT claims from request context
func GetClaims(r *http.Request) (*auth.Claims, error) {
	claims, ok := r.Context().Value(ContextKeyClaims).(*auth.Claims)
	if !ok {
		return nil, http.ErrNoCookie
	}
	return claims, nil
}
