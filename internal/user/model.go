package user

import (
	db "url-shortener/db/sqlc"
	"url-shortener/internal/utils"
)

type User struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	LastLoginAt string `json:"last_login_at"`
	CreatedAt   string `json:"created_at"`
}

func fromDBUser(dbUser db.User) User {
	return User{
		ID:          dbUser.ID,
		Email:       dbUser.Email,
		Password:    dbUser.Password,
		FirstName:   dbUser.FirstName.String,
		LastName:    dbUser.LastName.String,
		LastLoginAt: utils.ConvertTimeToString(dbUser.LastLoginAt.Time),
		CreatedAt:   utils.ConvertTimeToString(dbUser.CreatedAt),
	}
}
