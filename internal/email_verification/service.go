package emailverification

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
	db "url-shortener/db/sqlc"
	"url-shortener/internal/email"
	"url-shortener/internal/utils"
)

type Emailer interface {
	SendVerificationMail(email string, code string) error
	SendVerificationCompleteMail(email string) error
}

type EmailVerificationService struct {
	queries      *db.Queries
	emailService Emailer
}

func NewEmailVerificationService(queries *db.Queries, emailService Emailer) *EmailVerificationService {
	return &EmailVerificationService{
		queries:      queries,
		emailService: emailService,
	}
}

type StartEmailVerificationParams struct {
	UserID, Email string
}

func (s *EmailVerificationService) StartEmailVerification(ctx context.Context, args StartEmailVerificationParams) error {
	serviceID := "service.email_verification.StartEmailVerification"

	existingVerification, err := s.queries.GetEmailVerification(ctx, db.GetEmailVerificationParams{
		UserID: args.UserID,
		Email:  args.Email,
	})

	if err != nil {
		slog.Warn(serviceID, "message", "couldn't get email verification", "error", err)
	}

	if existingVerification.ExpiresAt.After(time.Now()) {
		if err = s.emailService.SendVerificationMail(existingVerification.Email, existingVerification.Code); err != nil {
			slog.Error(serviceID, "message", "couldn't send email verification", "error", err)
			return email.ErrSendingEmail
		}
		return nil
	}

	if err = s.queries.DeleteEmailVerification(ctx, existingVerification.ID); err != nil {
		slog.Warn(serviceID, "message", "couldn't delete email verification", "error", err)
	}

	code, err := utils.GenerateAlphanum(8)

	if err != nil {
		slog.Error(serviceID, "message", ErrGeneratingCode, "error", err)
		return ErrGeneratingCode
	}

	createEmailVerificationParams := db.CreateEmailVerificationParams{
		UserID:    args.UserID,
		Email:     args.Email,
		Code:      code,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	emailVerification, err := s.queries.CreateEmailVerification(ctx, createEmailVerificationParams)

	if err != nil {
		slog.Error(serviceID, "message", "couldn't create email verification", "error", err)
		return err
	}

	if err = s.emailService.SendVerificationMail(emailVerification.Email, emailVerification.Code); err != nil {
		slog.Error(serviceID, "message", "couldn't send email verification", "error", err)
		return email.ErrSendingEmail
	}

	slog.Info("email verification sent", "email", emailVerification.Email, "code", emailVerification.Code)

	return nil
}

type IsEmailVerifiedParams struct {
	UserID, Email string
}

func (s *EmailVerificationService) IsEmailVerified(ctx context.Context, args IsEmailVerifiedParams) (bool, error) {
	serviceID := "service.email_verification.IsEmailVerified"

	verified, err := s.queries.IsEmailVerificationComplete(ctx, db.IsEmailVerificationCompleteParams{
		UserID: args.UserID,
		Email:  args.Email,
	})

	if err != nil {
		slog.Error(serviceID, "message", "couldn't confirm email verification status", "error", err, "email", args.Email)
		return false, ErrUnknownError
	}

	if verified != 1 {
		return false, nil
	}

	return true, nil
}

type CompleteEmailVerificationParams struct {
	UserID, Email, Code string
}

func (s *EmailVerificationService) CompleteEmailVerification(ctx context.Context, args CompleteEmailVerificationParams) error {
	serviceID := "service.email_verification.CompleteEmailVerification"

	emailVerification, err := s.queries.CompleteEmailVerification(ctx, db.CompleteEmailVerificationParams{
		UserID: args.UserID,
		Email:  args.Email,
		Code:   args.Code,
	})

	if err != nil {
		if err == sql.ErrNoRows {
			slog.Error(serviceID, "message", "code is invalid", "error", err)
			return ErrInvalidVerificationCode
		}
		slog.Error(serviceID, "message", "couldn't verify email", "error", err)
		return ErrUnknownError
	}

	if err := s.emailService.SendVerificationCompleteMail(emailVerification.Email); err != nil {
		slog.Warn(serviceID, "message", "couldn't send verification complete mail")
	}

	return err
}
