package emailverification

import "errors"

var (
	ErrCreatingEmailVerification   = errors.New("couldn't create email verification")
	ErrRecreatingEmailVerification = errors.New("couldn't re-create email verification")
	ErrEmailVerificationCompleted  = errors.New("email verfiication already completed")
	ErrInvalidVerificationCode     = errors.New("invalid verification code")
	ErrUnknownError                = errors.New("something went wrong")
)
