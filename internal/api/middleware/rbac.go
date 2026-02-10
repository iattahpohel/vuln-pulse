package middleware

import (
	"net/http"

	"github.com/atta/vulnpulse/pkg/auth"
	"github.com/atta/vulnpulse/pkg/logger"
)

// RBACMiddleware checks if the user has the required permission
func RBACMiddleware(permission auth.Permission, log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := GetClaims(r)
			if err != nil {
				log.Warn("failed to get claims from context")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if err := auth.HasPermission(claims.Role, permission); err != nil {
				log.Warn("permission denied",
					"user_id", claims.UserID,
					"role", claims.Role,
					"permission", permission,
				)
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
