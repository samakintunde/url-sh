package server

import (
	"context"
	"fmt"
	"net/http"
	"url-shortener/internal/link"
)

type linkResponse struct {
	ID          string `json:"id"`
	OriginalURL string `json:"original_url"`
	ShortURLID  string `json:"short_url_id"`
	UpdatedAt   string `json:"last_login_at"`
	CreatedAt   string `json:"created_at"`
}

func newLinkResponse(link link.Link) linkResponse {
	return linkResponse{
		ID:          link.ID,
		OriginalURL: link.OriginalUrl,
		ShortURLID:  link.ShortUrlID,
		UpdatedAt:   link.UpdatedAt,
		CreatedAt:   link.CreatedAt,
	}
}

func HandleCreateShortLink(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello link")
	})
}
