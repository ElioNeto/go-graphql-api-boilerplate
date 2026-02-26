package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const ContextKeyUserID contextKey = "userID"

// JWTAuth middleware checks for a valid JWT in the Authorization header.
// It does NOT block requests without a token, as GraphQL relies on resolvers
// to determine if a specific field/query requires authentication.
func JWTAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				next.ServeHTTP(w, r)
				return
			}

			tokenStr := parts[1]
			claims := jwt.MapClaims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})

			if err == nil && token.Valid {
				// Inject the user ID into the request context
				if userIDStr, ok := claims["sub"].(string); ok {
					ctx := context.WithValue(r.Context(), ContextKeyUserID, userIDStr)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				} else if userIDFloat, ok := claims["sub"].(float64); ok {
					// JWT decodes numbers as float64
					ctx := context.WithValue(r.Context(), ContextKeyUserID, int64(userIDFloat))
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserIDFromContext retrieves the user ID from the context.
func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	val := ctx.Value(ContextKeyUserID)
	if id, ok := val.(int64); ok {
		return id, true
	}
	return 0, false
}
