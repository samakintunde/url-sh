package emailverification

import (
	"time"
	db "url-shortener/db/sqlc"
)

type EmailVerification struct {
	ID         int64
	Email      string
	VerifiedAt time.Time
	CreatedAt  time.Time
}

func fromDBEmailVerification(e db.EmailVerification) EmailVerification {
	return EmailVerification{
		ID:         e.ID,
		Email:      e.Email,
		VerifiedAt: e.VerifiedAt.Time,
		CreatedAt:  e.CreatedAt,
	}
}
