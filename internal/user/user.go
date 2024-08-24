package user

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
	db "url-shortener/db/sqlc"
	"url-shortener/internal/auth"
	"url-shortener/internal/token"
	"url-shortener/internal/utils"

	"github.com/mattn/go-sqlite3"
)

type UserService struct {
	queries    *db.Queries
	tokenMaker token.Maker
}

func NewUserService(queries *db.Queries, tokenMaker token.Maker) *UserService {
	return &UserService{
		queries:    queries,
		tokenMaker: tokenMaker,
	}
}

type CreateUserParams struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

func (u *UserService) CreateUser(ctx context.Context, args CreateUserParams) (db.User, error) {
	serviceID := "service.user.CreateUser"

	id := utils.NewULID()

	isPasswordStrong := auth.CheckPasswordStrength(args.Password)

	if !isPasswordStrong {
		slog.Info(serviceID, "error", ErrPasswordWeak)
		return db.User{}, ErrPasswordWeak
	}

	// Check if Password is being re-used or has been leaked
	isPasswordPwned, err := auth.CheckPasswordPwned(args.Password)

	if err != nil {
		slog.Info(serviceID, "message", "couldn't check if password pwned", "email", args.Email, "error", err)
		return db.User{}, ErrUnknownError
	}

	if isPasswordPwned {
		slog.Info(serviceID, "message", "password compromised in data breach", "email", args.Email)
		return db.User{}, ErrPasswordCompromised
	}

	// May fail because of the underlying rand.Read.
	// https://stackoverflow.com/a/42318347
	hashedPassword, err := auth.HashPassword(args.Password)

	if err != nil {
		slog.Error(serviceID, "error", ErrHashingPassword)
		return db.User{}, ErrHashingPassword
	}

	userParams := db.CreateUserParams{
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

	createdUser, err := u.queries.CreateUser(ctx, userParams)

	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			slog.Error(serviceID, "error", ErrUserExists)
			return db.User{}, ErrUserExists
		}
		return db.User{}, ErrCreatingUser
	}

	return createdUser, nil
}

type GetUserByIDParams struct {
	ID string
}

func (u *UserService) GetUserByID(ctx context.Context, args GetUserByIDParams) (db.User, error) {
	user, err := u.queries.GetUser(ctx, args.ID)

	if err != nil {
		return db.User{}, err
	}

	return user, nil
}

type GetUserByEmailParams struct {
	Email string
}

func (u *UserService) GetUserByEmail(ctx context.Context, args GetUserByEmailParams) (db.User, error) {
	serviceID := "service.user.GetUserByEmail"

	user, err := u.queries.GetUserByEmail(ctx, args.Email)

	if err != nil {
		if err == sql.ErrNoRows {
			slog.Error(serviceID, "message", "user not found", "email", args.Email, "error", err)
			return db.User{}, ErrUserNotFound
		}
		slog.Error(serviceID, "message", "database error querying existing user", "error", err)
		return db.User{}, err
	}

	return user, nil
}

type UpdateLoginTimeParams struct {
	ID   string
	Time time.Time
}

func (u *UserService) UpdateLastLogin(ctx context.Context, args UpdateLoginTimeParams) error {
	serviceID := "service.user.UpdateLastLogin"

	err := u.queries.UpdateUserLoginTime(ctx, db.UpdateUserLoginTimeParams{
		ID: args.ID,
		LastLoginAt: sql.NullTime{
			Time:  args.Time,
			Valid: !args.Time.IsZero(),
		},
	})

	if err != nil {
		slog.Error(serviceID, "message", "database error updating user login time", "error", err)
	}

	return err
}

type StartPasswordResetParams struct {
	Email string
}

func (u *UserService) StartPasswordReset(ctx context.Context, args StartPasswordResetParams) (db.PasswordResetToken, error) {
	serviceID := "service.user.StartPasswordReset"

	user, err := u.queries.GetUserByEmail(ctx, args.Email)

	if err != nil {
		if err == sql.ErrNoRows {
			slog.Error(serviceID, "message", "user not found", "email", args.Email, "error", err)
			return db.PasswordResetToken{}, ErrUserNotFound
		}
		slog.Error(serviceID, "message", "database error querying existing user", "error", err)
		return db.PasswordResetToken{}, err
	}

	existingPasswordResetToken, err := u.queries.GetPasswordResetTokenByUserID(ctx, user.ID)

	if err != nil {
		if err == sql.ErrNoRows {
			slog.Error(serviceID, "message", "no valid password reset found", "email", args.Email, "error", err)
		} else {
			slog.Error(serviceID, "message", "database error querying existing password token", "error", err)
			return db.PasswordResetToken{}, ErrUnknownError
		}
	}

	if existingPasswordResetToken.Token != "" && existingPasswordResetToken.ExpiresAt.After(time.Now()) {
		return existingPasswordResetToken, nil
	}

	token, claims, err := u.tokenMaker.CreateToken(user.ID, 1*time.Hour)

	if err != nil {
		slog.Error(serviceID, "message", "couldn't create token", "email", args.Email, "error", err)
		return db.PasswordResetToken{}, ErrUnknownError
	}

	createPasswordResetParams := db.CreatePasswordResetTokenParams{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: claims.ExpiresAt,
	}

	passwordResetToken, err := u.queries.CreatePasswordResetToken(ctx, createPasswordResetParams)

	if err != nil {
		slog.Error(serviceID, "message", "couldn't create start password reset", "email", args.Email, "error", err)
		return db.PasswordResetToken{}, err
	}

	return passwordResetToken, nil
}

type ResetPasswordParams struct {
	Token    string
	Password string
}

func (u *UserService) ResetPassword(ctx context.Context, args ResetPasswordParams) (db.PasswordResetToken, error) {
	serviceID := "service.user.ResetPassword"

	_, err := u.tokenMaker.VerifyToken(args.Token)

	if err != nil {
		return db.PasswordResetToken{}, err
	}

	passwordResetToken, err := u.queries.GetPasswordResetToken(ctx, args.Token)

	if err != nil {
		if err == sql.ErrNoRows {
			slog.Error(serviceID, "message", "password reset not started", "user", passwordResetToken.UserID, "error", err)
			return db.PasswordResetToken{}, ErrUserNotFound
		}
		slog.Error(serviceID, "message", "couldn't get password reset token", "user", passwordResetToken.UserID, "error", err)
		return db.PasswordResetToken{}, err
	}

	isPasswordStrong := auth.CheckPasswordStrength(args.Password)

	if !isPasswordStrong {
		slog.Info(serviceID, "error", ErrPasswordWeak)
		return db.PasswordResetToken{}, ErrPasswordWeak
	}

	// Check if Password is being re-used or has been leaked
	isPasswordPwned, err := auth.CheckPasswordPwned(args.Password)

	if err != nil {
		slog.Info(serviceID, "message", "couldn't check if password pwned", "user", passwordResetToken.ID, "error", err)
		return db.PasswordResetToken{}, ErrUnknownError
	}

	if isPasswordPwned {
		slog.Info(serviceID, "message", "password compromised in data breach", "user", passwordResetToken.ID)
		return db.PasswordResetToken{}, ErrPasswordCompromised
	}

	hashedPassword, err := auth.HashPassword(args.Password)

	if err != nil {
		slog.Error(serviceID, "message", "couldn't hash password", "user", passwordResetToken.UserID, "error", err)
		return db.PasswordResetToken{}, ErrHashingPassword
	}

	user, err := u.queries.GetUser(ctx, passwordResetToken.UserID)

	if err != nil {
		slog.Error(serviceID, "message", "couldn't get user", "user", passwordResetToken.UserID, "error", err)
		return db.PasswordResetToken{}, ErrGettingUserByID
	}

	isCurrentPassword, err := auth.VerifyPassword(args.Password, user.Password)

	if isCurrentPassword {
		slog.Error(serviceID, "message", "current password being reused", "user", passwordResetToken.UserID, "error", err)
		return db.PasswordResetToken{}, ErrReusingPassword
	}

	err = u.queries.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		Password: hashedPassword,
		ID:       passwordResetToken.UserID,
	})

	if err != nil {
		slog.Error(serviceID, "message", "couldn't update user password", "user", passwordResetToken.UserID, "error", err)
		return db.PasswordResetToken{}, err
	}

	u.queries.DeletePasswordResetToken(ctx, passwordResetToken.Token)

	return passwordResetToken, nil
}
