package token

import (
	"time"
)

type Maker interface {
	// CreateToken creates a new token for a specific user id and duration
	CreateToken(userID string, duration time.Duration) (string, *Claims, error)
	// VerifyToken checks if the token is valid or not
	VerifyToken(token string) (*Claims, error)
}
