package server

import (
	"context"
	"net/http"
	db "url-shortener/db/sqlc"
	"url-shortener/internal/analytics"
	"url-shortener/internal/auth"
	"url-shortener/internal/config"
	"url-shortener/internal/email"
	emailverification "url-shortener/internal/email_verification"
	"url-shortener/internal/link"
	"url-shortener/internal/token"
	"url-shortener/internal/user"
	"url-shortener/internal/validation"
)

func New(ctx context.Context, cfg config.Config, fs http.Handler, queries *db.Queries, tokenMaker token.Maker) http.Handler {
	mux := http.NewServeMux()

	validator := validation.NewValidationService()
	var emailService email.Emailer
	// if cfg.Debug {
	emailService = email.NewMockEmailService()
	// } else {
	// 	emailService = email.NewResendService(email.EmailSMTPConfig(cfg.SMTP), cfg.ResendKey)
	// }
	emailVerificationService := emailverification.NewEmailVerificationService(queries, emailService)
	userService := user.NewUserService(queries, tokenMaker, emailService, emailVerificationService)
	authService := auth.NewAuthService(queries)
	linkService := link.NewLinkService(queries)
	analyticsService := analytics.NewAnalyticsService(queries)

	routes(ctx, mux, fs, validator, tokenMaker, userService, authService, emailVerificationService, linkService, analyticsService)
	return mux
}
