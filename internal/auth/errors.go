package auth

import "errors"

var (
	ErrWeakPassword           = errors.New("weak_password")
	ErrCompromisedPassword    = errors.New("compromised_password")
	ErrHashingPassword        = errors.New("hashing_password_failed")
	ErrDecodingHashedPassword = errors.New("decoding_hashed_password_failed")
)
