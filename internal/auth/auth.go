package auth

import db "url-shortener/db/sqlc"

type AuthService struct {
	queries *db.Queries
}

func NewAuthService(queries *db.Queries) *AuthService {
	return &AuthService{
		queries: queries,
	}
}
