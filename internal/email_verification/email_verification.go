package emailverification

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"
	db "url-shortener/db/sqlc"
	"url-shortener/internal/email"
	"url-shortener/internal/utils"
)

type Emailer interface {
	Send(to []string, subject, body string) error
}

type EmailVerificationService struct {
	queries *db.Queries
	emailer Emailer
}

func NewEmailVerificationService(queries *db.Queries, emailer Emailer) *EmailVerificationService {
	return &EmailVerificationService{
		queries: queries,
		emailer: emailer,
	}
}

type CreateEmailVerificationParams struct {
	UserID, UserEmail string
}

func (e *EmailVerificationService) CreateEmailVerification(ctx context.Context, args CreateEmailVerificationParams) error {
	serviceID := "service.email_verification.CreateEmailVerification"

	code, err := utils.GenerateAlphanum(8)

	if err != nil {
		slog.Error(serviceID, "message", utils.ErrGeneratingToken, "error", err)
		return utils.ErrGeneratingToken
	}

	emailVerificationParam := db.CreateEmailVerificationParams{
		UserID:    args.UserID,
		Email:     args.UserEmail,
		Code:      code,
		ExpiresAt: time.Now().Add(time.Minute * 15),
	}

	err = e.queries.CreateEmailVerification(ctx, emailVerificationParam)

	if err != nil {
		slog.Error(serviceID, "message", ErrCreatingEmailVerification, "error", err)
		return ErrCreatingEmailVerification
	}

	slog.Info(serviceID, "message", "created email verification", "email", args.UserEmail, "code", code)

	err = e.emailer.Send([]string{args.UserEmail}, "Verify your Account", fmt.Sprintf("Your verification code is: %s", code))

	if err != nil {
		slog.Error(serviceID, "message", "couldn't send email verification", "error", err)
		return email.ErrSendingEmail
	}

	return nil
}

type RecreateEmailVerificationParams CreateEmailVerificationParams

func (e *EmailVerificationService) RecreateEmailVerification(ctx context.Context, args CreateEmailVerificationParams) (db.GetEmailVerificationRow, error) {
	serviceID := "service.email_verification.RereateEmailVerification"

	code, err := utils.GenerateAlphanum(8)

	if err != nil {
		slog.Error(serviceID, "message", utils.ErrGeneratingToken, "error", err)
		return db.GetEmailVerificationRow{}, utils.ErrGeneratingToken
	}

	emailVerification, err := e.queries.GetEmailVerification(ctx, db.GetEmailVerificationParams{
		UserID: args.UserID,
		Email:  args.UserEmail,
	})

	if err != nil {
		if err == sql.ErrNoRows {
			slog.Info(serviceID, "message", "no email verification found", "error", err, "email", args.UserEmail)
		} else {
			return db.GetEmailVerificationRow{}, err
		}
	}

	if emailVerification.ExpiresAt.After(time.Now()) {
		recreatedEmailVerification, err := e.queries.RecreateEmailVerification(ctx, db.RecreateEmailVerificationParams{
			Code:      emailVerification.Code,
			ExpiresAt: time.Now().Add(time.Minute * 15),
			UserID:    emailVerification.UserID,
			Email:     emailVerification.Email,
		})
		if err != nil {
			slog.Error(serviceID, "message", ErrRecreatingEmailVerification, "error", err)
			return db.GetEmailVerificationRow{}, ErrCreatingEmailVerification

		}
		return db.GetEmailVerificationRow(recreatedEmailVerification), nil
	} else {
		e.queries.CleanExpiredEmailVerificationsForUserID(ctx, emailVerification.UserID)
	}

	emailVerificationParam := db.CreateEmailVerificationParams{
		UserID:    args.UserID,
		Email:     args.UserEmail,
		Code:      code,
		ExpiresAt: time.Now().Add(time.Minute * 15),
	}

	err = e.queries.CreateEmailVerification(ctx, emailVerificationParam)

	if err != nil {
		slog.Error(serviceID, "message", ErrCreatingEmailVerification, "error", err)
		return db.GetEmailVerificationRow{}, ErrCreatingEmailVerification
	}

	slog.Info(serviceID, "message", "created email verification", "email", args.UserEmail, "code", code)

	err = e.emailer.Send([]string{args.UserEmail}, "Verify your Account", fmt.Sprintf("Your verification code is: %s", code))

	if err != nil {
		slog.Error(serviceID, "message", "couldn't send email verification", "error", err)
		return db.GetEmailVerificationRow{}, email.ErrSendingEmail
	}

	return emailVerification, nil
}

type HasUserCompletedEmailVerificationParams struct {
	UserID, UserEmail string
}

func (e *EmailVerificationService) HasUserCompletedEmailVerification(ctx context.Context, args HasUserCompletedEmailVerificationParams) (bool, error) {
	serviceID := "service.email_verification.HasUserCompletedEmailVerification"

	hasCompletedVerification, err := e.queries.IsUserEmailVerificationComplete(ctx, db.IsUserEmailVerificationCompleteParams{
		UserID: args.UserID,
		Email:  args.UserEmail,
	})

	if err != nil {
		slog.Error(serviceID, "message", "couldn't confirm email verification status", "error", err, "email", args.UserEmail)
		return false, ErrUnknownError
	}

	if hasCompletedVerification == 1 {
		slog.Info(serviceID, "message", "email already verified")
		return true, nil
	}

	return false, nil
}

type CompleteEmailVerificationParams struct {
	UserID, UserEmail, Code string
}

func (e *EmailVerificationService) CompleteEmailVerification(ctx context.Context, args CompleteEmailVerificationParams) error {
	serviceID := "service.email_verification.CompleteEmailVerification"

	_, err := e.queries.CompleteEmailVerification(ctx, db.CompleteEmailVerificationParams{
		UserID: args.UserID,
		Email:  args.UserEmail,
		Code:   args.Code,
	})

	if err != nil {
		if err == sql.ErrNoRows {
			slog.Error(serviceID, "message", "code is invalid", "error", err, "email", args.UserEmail, "code", args.Code)
			return ErrInvalidVerificationCode
		}
		slog.Error(serviceID, "message", "couldn't verify email", "error", err, "email", args.UserEmail)
		return ErrUnknownError
	}

	return nil
}
