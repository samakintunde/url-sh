package user

import "errors"

var (
	ErrPasswordWeak        = errors.New("password is too weak")
	ErrPasswordCompromised = errors.New("password exposed in data breach")
	ErrHashingPassword     = errors.New("couldn't hash password")
	ErrReusingPassword     = errors.New("can't reuse current password")
	ErrUserExists          = errors.New("user already exists")
	ErrCreatingUser        = errors.New("error creating user")
	ErrUserNotFound        = errors.New("user not found")
	ErrGettingUserByID     = errors.New("error getting user by ID")
	ErrGettingUserByEmail  = errors.New("error getting user by email")
	ErrUnknownError        = errors.New("something went wrong")
)
