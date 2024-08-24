package server

import (
	"context"
	"net/http"
	db "url-shortener/db/sqlc"
	"url-shortener/internal/auth"
	"url-shortener/internal/email"
	emailverification "url-shortener/internal/email_verification"
	"url-shortener/internal/token"
	"url-shortener/internal/user"
	"url-shortener/internal/validation"
)

func New(ctx context.Context, fs http.Handler, queries *db.Queries, validator validation.Validator, emailer email.Emailer, tokenMaker token.Maker) http.Handler {
	mux := http.NewServeMux()

	userService := user.NewUserService(queries, tokenMaker)
	authService := auth.NewAuthService(queries)
	emailVerificationService := emailverification.NewEmailVerificationService(queries, emailer)

	routes(ctx, mux, fs, validator, emailer, tokenMaker, userService, authService, emailVerificationService)
	return mux
}
