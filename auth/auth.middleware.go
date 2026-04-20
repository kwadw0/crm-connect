package auth

import (
	"context"
	"kwadw0/WhatsCRM/auth/jwt"
	"kwadw0/WhatsCRM/utils"
	"net/http"
	"strings"
)

// To keep thing type-safe, we create a private key type for our context.
type contextKey string

const userIDKey contextKey = "userID"

// AuthMiddleware is our "Guard" for protected routes.
func AuthMiddleware(jwtSecret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Extract the Token from the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				utils.WriteJson(w, http.StatusUnauthorized, "Missing or invalid token", nil, "Bearer token required")
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			// 2. Validate the Token
			token, err := authjwt.ValidateToken(jwtSecret, tokenString)
			if err != nil || !token.Valid {
				utils.WriteJson(w, http.StatusUnauthorized, "Invalid token", nil, "Your session has expired or the token is corrupt")
				return
			}

			// 3. Extract the Claims (The Payload)
			claims, ok := token.Claims.(*authjwt.CustomClaims)
			if !ok {
				utils.WriteJson(w, http.StatusUnauthorized, "Invalid claims", nil, "Failed to parse User details from token")
				return
			}

			// 4. Inject into Context (the "Pocket")
			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			
			// 5. Pass the request with the new context to the next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserIDFromContext is a helper for any handler to easily get the logged in User ID.
func GetUserIDFromContext(ctx context.Context) string {
	userID, _ := ctx.Value(userIDKey).(string)
	return userID
}
