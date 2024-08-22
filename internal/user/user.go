package user

import (
	"context"
	"database/sql"
	"log/slog"
	db "url-shortener/db/sqlc"
	"url-shortener/internal/auth"
	"url-shortener/internal/utils"

	"github.com/mattn/go-sqlite3"
)

type UserService struct {
	queries *db.Queries
}

func NewUserService(queries *db.Queries) *UserService {
	return &UserService{
		queries: queries,
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
