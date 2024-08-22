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

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

func New(ctx context.Context, fs http.Handler, queries *db.Queries, validate *validator.Validate, ut *ut.UniversalTranslator, trans ut.Translator, emailer email.Emailer, tokenMaker token.Maker) http.Handler {
	mux := http.NewServeMux()

	userService := user.NewUserService(queries)
	authService := auth.NewAuthService(queries)
	emailVerificationService := emailverification.NewEmailVerificationService(queries, emailer)

	routes(ctx, mux, fs, validate, ut, trans, emailer, tokenMaker, userService, authService, emailVerificationService)
	return mux
}
