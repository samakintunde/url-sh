package server

import (
	"context"
	"fmt"
	"net/http"
	"url-shortener/internal/auth"
	"url-shortener/internal/email"
	emailverification "url-shortener/internal/email_verification"
	"url-shortener/internal/token"
	"url-shortener/internal/user"
	"url-shortener/internal/validation"
)

func routes(ctx context.Context, mux *http.ServeMux, fs http.Handler, validator validation.Validator, emailer email.Emailer, tokenMaker token.Maker, userService *user.UserService, _ *auth.AuthService, emailVerificationService *emailverification.EmailVerificationService) {
	// AUTH
	apiMux := NewRouteGroup("/api", mux)
	apiMux.Handle("POST /auth/signup", HandleSignup(ctx, validator, *userService, *emailVerificationService))
	apiMux.Handle("POST /auth/email-verification", HandleVerifyEmail(ctx, validator, *userService, *emailVerificationService))
	apiMux.Handle("POST /auth/login", HandleLogin(ctx, validator, tokenMaker, userService, emailVerificationService, emailer))
	apiMux.Handle("POST /auth/password-reset/start", HandleStartResetPassword(ctx, validator, userService, emailer))
	apiMux.Handle("POST /auth/password-reset", HandleResetPassword(ctx, validator, userService, emailer))

	// PROFILE
	userMux := apiMux.Group("/user")
	userMux.Use(VerifyAuth(tokenMaker))
	userMux.Handle("GET /me", handleNoop())
	userMux.Handle("GET /change-password", handleNoop())
	// mux.HandleFunc("GET /api/", VerifyAuth(func(w http.ResponseWriter, r *http.Request) {
	// 	w.Write([]byte("Cooking..."))
	// }, tokenMaker))
	//

	mux.Handle("GET /", fs)
	mux.HandleFunc("GET /health/", HandleHealth())
}

func handleNoop() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, ".")
	})
}
