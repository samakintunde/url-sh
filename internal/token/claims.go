package token

import (
	"errors"
	"time"
)

var (
	ErrExpiredToken = errors.New("token has expired")
)

type Claims struct {
	UserID    string    `json:"user_id"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

func NewClaims(userID string, duration time.Duration) *Claims {
	return &Claims{
		UserID:    userID,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(duration),
	}
}

// Valid checks if the token payload is valid or not
func (claims *Claims) Valid() error {
	if time.Now().After(claims.ExpiresAt) {
		return ErrExpiredToken
	}
	return nil
}
