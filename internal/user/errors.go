package user

import "errors"

var (
	ErrReusingPassword            = errors.New("can't reuse current password")
	ErrUserExists                 = errors.New("user already exists")
	ErrCreatingUser               = errors.New("error creating user")
	ErrUserNotFound               = errors.New("user not found")
	ErrPasswordResetTokenNotFound = errors.New("password reset token not found")
	ErrGettingUserByID            = errors.New("error getting user by ID")
	ErrGettingUserByEmail         = errors.New("error getting user by email")
	ErrInvalidVerificationToken   = errors.New("password reset token is invalid")
	ErrInvalidPasswordResetToken  = errors.New("password reset token is invalid")
	ErrEmailVerificationRequired  = errors.New("email_verification_required")
	ErrUnknownError               = errors.New("something went wrong")
)
