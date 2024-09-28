package user

import (
	"context"
	"database/sql"
	"encoding/base64"
	"log/slog"
	"time"
	db "url-shortener/db/sqlc"

	"url-shortener/internal/auth"
	"url-shortener/internal/email"
	emailverification "url-shortener/internal/email_verification"
	"url-shortener/internal/token"
	"url-shortener/internal/utils"

	"github.com/mattn/go-sqlite3"
)

type Emailer interface {
	SendPasswordResetMail(email string, token string) error
}

type UserService struct {
	queries                  *db.Queries
	tokenMaker               token.Maker
	emailVerificationService *emailverification.EmailVerificationService
	emailService             Emailer
}

// Cyclic dependencies
func NewUserService(queries *db.Queries, tokenMaker token.Maker, emailService Emailer, emailVerificationService *emailverification.EmailVerificationService) *UserService {
	return &UserService{
		queries:                  queries,
		tokenMaker:               tokenMaker,
		emailService:             emailService,
		emailVerificationService: emailVerificationService,
	}
}

type RegisterUserParams struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

func (u *UserService) RegisterUser(ctx context.Context, args RegisterUserParams) (User, error) {
	const serviceID = "service.user.RegisterUser"

	id := utils.NewULID()

	err := auth.CheckPassword(args.Password)

	if err != nil {
		slog.Info(serviceID, "error", err)
		return User{}, err
	}

	hashedPassword, err := auth.HashPassword(args.Password)

	if err != nil {
		slog.Error(serviceID, "error", auth.ErrHashingPassword)
		return User{}, auth.ErrHashingPassword
	}

	createUserParams := db.CreateUserParams{
		ID:       id.String(),
		Email:    args.Email,
		Password: hashedPassword,
		FirstName: sql.NullString{
			String: args.FirstName,
			Valid:  args.FirstName != "",
		},
		LastName: sql.NullString{
			String: args.LastName,
			Valid:  args.LastName != "",
		},
	}

	createdUser, err := u.queries.CreateUser(ctx, createUserParams)

	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			slog.Error(serviceID, "error", ErrUserExists)
			return User{}, ErrUserExists
		}
		return User{}, ErrCreatingUser
	}

	startEmailVerificationParams := emailverification.StartEmailVerificationParams{
		UserID: createdUser.ID,
		Email:  createdUser.Email,
	}
	err = u.emailVerificationService.StartEmailVerification(ctx, startEmailVerificationParams)

	if err != nil {
		// TODO: add retry mechanism
		slog.Warn("failed to start email verification", "error", err)
	}

	return fromDBUser(createdUser), nil
}

type LoginUserParams struct {
	Email    string
	Password string
}

func (s *UserService) LoginUser(ctx context.Context, args LoginUserParams) (User, error) {
	const serviceID = "service.user.LoginUser"

	user, err := s.queries.GetUserByEmail(ctx, args.Email)

	if err != nil {
		slog.Error(serviceID, "message", "database error querying existing user", "error", err)
		return User{}, err
	}

	doPasswordsMatch, err := auth.VerifyPassword(args.Password, user.Password)

	if err != nil {
		slog.Error(serviceID, "message", "Error verifying password", "error", err)
		return User{}, err
	}

	if !doPasswordsMatch {
		slog.Error(serviceID, "message", "incorrect password")
		return User{}, err
	}

	hasCompletedEmailVerificationArgs := emailverification.IsEmailVerifiedParams{
		UserID: user.ID,
		Email:  user.Email,
	}

	isVerified, err := s.emailVerificationService.IsEmailVerified(ctx, hasCompletedEmailVerificationArgs)

	if err != nil {
		slog.Error(serviceID, "message", "couldn't confirm email verification status", "error", err)
		return User{}, err
	}

	if !isVerified {
		slog.Info(serviceID, "message", "email not verified")
		args := emailverification.StartEmailVerificationParams{
			UserID: user.ID,
			Email:  user.Email,
		}

		if err := s.emailVerificationService.StartEmailVerification(ctx, args); err != nil {
			slog.Warn(serviceID, "message", "couldn't start email verification", "error", err)
		}

		return User{}, ErrEmailVerificationRequired
	}

	updateLoginTimeArgs := db.UpdateUserLoginTimeParams{
		ID: user.ID,
		LastLoginAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
	}

	if err = s.queries.UpdateUserLoginTime(ctx, updateLoginTimeArgs); err != nil {
		slog.Error(serviceID, "message", "couldn't update login time", "error", err)
		return User{}, err
	}

	user.LastLoginAt = updateLoginTimeArgs.LastLoginAt

	return fromDBUser(user), nil
}

type StartPasswordResetParams struct {
	Email string
}

func (s *UserService) StartPasswordReset(ctx context.Context, args StartPasswordResetParams) error {
	serviceID := "service.user.StartPasswordReset"

	user, err := s.queries.GetUserByEmail(ctx, args.Email)

	if err != nil {
		if err == sql.ErrNoRows {
			slog.Error(serviceID, "message", "user not found", "error", err)
			return ErrUserNotFound
		}
		slog.Error(serviceID, "message", "database error querying existing user", "error", err)
		return ErrUnknownError
	}

	existingPasswordResetToken, err := s.queries.GetPasswordResetTokenByUserID(ctx, user.ID)

	if err != nil {
		if err == sql.ErrNoRows {
			slog.Error(serviceID, "message", "no valid password reset found", "email", args.Email, "error", err)
		} else {
			slog.Error(serviceID, "message", "database error querying existing password token", "error", err)
			return ErrUnknownError
		}
	}

	if existingPasswordResetToken.ExpiresAt.After(time.Now()) {
		if err := s.emailService.SendPasswordResetMail(user.Email, existingPasswordResetToken.Token); err != nil {
			slog.Error(serviceID, "message", "couldn't send password reset mail", "error", err)
			return email.ErrSendingEmail
		}
		return nil
	}

	if err := s.queries.DeletePasswordResetToken(ctx, existingPasswordResetToken.ID); err != nil {
		slog.Error(serviceID, "message", "couldn't delete password reset token", "error", err)
	}

	token, claims, err := s.tokenMaker.CreateToken(user.ID, 1*time.Hour)

	if err != nil {
		slog.Error(serviceID, "message", "couldn't create token", "email", args.Email, "error", err)
		return ErrUnknownError
	}

	createPasswordResetParams := db.CreatePasswordResetTokenParams{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: claims.ExpiresAt,
	}

	passwordResetToken, err := s.queries.CreatePasswordResetToken(ctx, createPasswordResetParams)

	if err != nil {
		slog.Error(serviceID, "message", "couldn't create start password reset", "email", args.Email, "error", err)
		return ErrUnknownError
	}

	encodedToken := base64.RawStdEncoding.EncodeToString([]byte(passwordResetToken.Token))

	if err := s.emailService.SendPasswordResetMail(user.Email, encodedToken); err != nil {
		slog.Error(serviceID, "message", "couldn't send password reset mail", "error", err)
		return email.ErrSendingEmail
	}

	return nil
}

type ResetPasswordParams struct {
	Token    string
	Password string
}

func (u *UserService) ResetPassword(ctx context.Context, args ResetPasswordParams) error {
	serviceID := "service.user.ResetPassword"

	decodedToken, err := base64.RawStdEncoding.DecodeString(args.Token)

	token := string(decodedToken)

	if err != nil {
		slog.Error(serviceID, "message", "couldn't decode base64 token", "error", err)
		return ErrInvalidPasswordResetToken
	}

	claims, err := u.tokenMaker.VerifyToken(token)

	if err != nil {
		slog.Error(serviceID, "message", "couldn't verify token", "error", err)
		return ErrInvalidPasswordResetToken
	}

	userID := claims.UserID

	if userID == "" {
		return ErrInvalidPasswordResetToken
	}

	passwordResetToken, err := u.queries.GetPasswordResetToken(ctx, db.GetPasswordResetTokenParams{
		UserID: userID,
		Token:  token,
	})

	if err != nil {
		if err == sql.ErrNoRows {
			slog.Error(serviceID, "message", "password reset not started", "user", passwordResetToken.UserID, "error", err)
			return ErrPasswordResetTokenNotFound
		}
		slog.Error(serviceID, "message", "couldn't get password reset token", "error", err)
		return ErrUnknownError
	}

	err = auth.CheckPassword(args.Password)

	if err != nil {
		slog.Info(serviceID, "message", auth.ErrWeakPassword, "error", err)
		return err
	}

	hashedPassword, err := auth.HashPassword(args.Password)

	if err != nil {
		slog.Error(serviceID, "message", "couldn't hash password", "user", passwordResetToken.UserID, "error", err)
		return auth.ErrHashingPassword
	}

	user, err := u.queries.GetUser(ctx, passwordResetToken.UserID)

	if err != nil {
		slog.Error(serviceID, "message", "couldn't get user", "user", passwordResetToken.UserID, "error", err)
		return ErrGettingUserByID
	}

	isCurrentPassword := hashedPassword == user.Password

	if isCurrentPassword {
		slog.Error(serviceID, "message", "current password being reused", "error", err)
		return ErrReusingPassword
	}

	updatePasswordParams := db.UpdateUserPasswordParams{
		Password: hashedPassword,
		ID:       passwordResetToken.UserID,
	}

	if err = u.queries.UpdateUserPassword(ctx, updatePasswordParams); err != nil {
		slog.Error(serviceID, "message", "couldn't update user password", "error", err)
		return ErrUnknownError
	}

	return u.queries.DeletePasswordResetToken(ctx, passwordResetToken.ID)
}

type VerifyEmailParams struct {
	Code string
}

func (u *UserService) VerifyEmail(ctx context.Context, args VerifyEmailParams) error {
	serviceID := "service.user.VerifyEmail"

	emailVerification, err := u.queries.GetEmailVerificationByCode(ctx, args.Code)

	if err != nil {
		slog.Error(serviceID, "message", "couldn't get email verification", "error", err)
		return ErrInvalidPasswordResetToken
	}

	err = u.emailVerificationService.CompleteEmailVerification(ctx, emailverification.CompleteEmailVerificationParams{
		UserID: emailVerification.UserID,
		Email:  emailVerification.Email,
		Code:   args.Code,
	})

	if err != nil {
		slog.Error(serviceID, "message", "couldn't update email verification", "error", err)
		return ErrUnknownError
	}

	return nil
}
