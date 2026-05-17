package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/stdpmk/hotels/internal/http/response"
)

type contextKey string

const UserIDKey contextKey = "userID"

func Auth(redis *redis.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				response.WriteError(w, http.StatusUnauthorized, "unauthorized", response.CodeUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")

			userID, err := redis.Get(r.Context(), "session:"+token).Int64()
			if err != nil {
				response.WriteError(w, http.StatusUnauthorized, "unauthorized", response.CodeUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
