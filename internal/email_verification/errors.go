package emailverification

import "errors"

var (
	ErrGeneratingCode             = errors.New("failed to generate verification code")
	ErrCreatingEmailVerification  = errors.New("couldn't create email verification")
	ErrExtendingEmailVerification = errors.New("couldn't extend email verification")
	ErrEmailVerificationStarted   = errors.New("email verfiication already started")
	ErrEmailVerificationCompleted = errors.New("email verfiication already completed")
	ErrInvalidVerificationCode    = errors.New("invalid verification code")
	ErrUnknownError               = errors.New("something went wrong")
)
