package server

import (
	"context"
	"net/http"
	"url-shortener/internal/auth"
	"url-shortener/internal/email"
	emailverification "url-shortener/internal/email_verification"
	"url-shortener/internal/token"
	"url-shortener/internal/user"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

func routes(ctx context.Context, mux *http.ServeMux, fs http.Handler, validate *validator.Validate, ut *ut.UniversalTranslator, trans ut.Translator, _ email.Emailer, tokenMaker token.Maker, userService *user.UserService, _ *auth.AuthService, emailVerificationService *emailverification.EmailVerificationService) {
	mux.Handle("POST /api/auth/signup", HandleSignup(ctx, *userService, *emailVerificationService, validate, trans))
	mux.Handle("POST /api/auth/email-verification", HandleVerifyEmail(ctx, *userService, *emailVerificationService, validate, ut, trans))
	mux.Handle("POST /api/auth/login", HandleLogin(ctx, *userService, validate, ut, trans, tokenMaker))
	mux.HandleFunc("GET /api/", VerifyAuth(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Cooking..."))
	}, tokenMaker))

	mux.Handle("GET /", fs)
	mux.HandleFunc("GET /health/", HandleHealth())
}
