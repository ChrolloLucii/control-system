package middleware

import (
	"context"
	"net/http"
	"strings"
	"user-service/internal/dto"
	"user-service/internal/service"

	"encoding/json"
)

type contextKey string

const UserContextKey contextKey = "user"

func AuthMiddleware(jwtService service.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authorization header required")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid authorization header format")
				return
			}

			token := parts[1]
			claims, err := jwtService.ValidateToken(token)
			if err != nil {
				respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(UserContextKey).(*service.Claims)
		if !ok {
			respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "user not authenticated")
			return
		}
		hasAdminRole := false
		for _, role := range claims.Roles {
			if role == "admin" {
				hasAdminRole = true
				break
			}
		}

		if !hasAdminRole {
			respondWithError(w, http.StatusForbidden, "FORBIDDEN", "admin access required")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func respondWithError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(dto.Response{
		Success: false,
		Error: &dto.ErrorDTO{
			Code:    code,
			Message: message,
		},
	})
}
