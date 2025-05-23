package link

import (
	db "url-shortener/db/sqlc"
	"url-shortener/internal/utils"
)

type Link struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	OriginalUrl string `json:"original_url"`
	ShortUrlID  string `json:"short_url_id"`
	PrettyID    string `json:"pretty_id"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
}

func fromDBLink(dbUser db.Link) Link {
	return Link{
		ID:          dbUser.ID,
		UserID:      dbUser.UserID,
		OriginalUrl: dbUser.OriginalUrl,
		ShortUrlID:  dbUser.ShortUrlID,
		PrettyID:    dbUser.PrettyID,
		UpdatedAt:   utils.ConvertTimeToString(dbUser.UpdatedAt),
		CreatedAt:   utils.ConvertTimeToString(dbUser.CreatedAt),
	}
}
