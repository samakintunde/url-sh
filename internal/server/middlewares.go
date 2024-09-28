package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"url-shortener/internal/token"
)

var (
	ErrMissingToken = errors.New("missing access token")
)

func VerifyAuth(tokenMaker token.Maker) func(http.Handler) http.Handler {
	middlewareID := "middleware.VerifyAuth"
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString, err := extractToken(r)

			if err != nil {
				slog.Error(middlewareID, "error", err)
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			claims, err := tokenMaker.VerifyToken(tokenString)

			if err != nil {
				slog.Error(middlewareID, "error", err)
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			if claims.UserID == "" {
				slog.Error(middlewareID, "error", "no data in claims")
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, "user_id", claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractToken(r *http.Request) (string, error) {
	accessTokenCookie, err := r.Cookie("access_token")
	var tokenString string
	if err != nil {
		tokenString = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	} else {
		tokenString = accessTokenCookie.Value
	}

	if tokenString == "" {
		return "", ErrMissingToken
	}

	return tokenString, nil

}
