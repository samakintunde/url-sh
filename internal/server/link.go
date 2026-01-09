package server

import (
	"context"
	"log/slog"
	"net/http"
	"url-shortener/internal/analytics"
	"url-shortener/internal/link"
	"url-shortener/internal/utils"
	"url-shortener/internal/validation"
)

type linkResponse struct {
	ID          string `json:"id"`
	OriginalURL string `json:"original_url"`
	ShortURLID  string `json:"short_url_id"`
	PrettyID    string `json:"pretty_id"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
}

func newLinkResponse(link link.Link) linkResponse {
	return linkResponse{
		ID:          link.ID,
		OriginalURL: link.OriginalUrl,
		ShortURLID:  link.ShortUrlID,
		PrettyID:    link.PrettyID,
		UpdatedAt:   link.UpdatedAt,
		CreatedAt:   link.CreatedAt,
	}
}

func HandleCreateShortLink(ctx context.Context, validator validation.Validator, linkService *link.LinkService) http.Handler {
	type request struct {
		OriginalURL string `json:"original_url" validate:"required,url"`
		PrettyID    string `json:"pretty_id"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(userIDKey).(string)
		if !ok {
			utils.RespondWithJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			return
		}

		req, err := utils.DecodeToJSON[request](r)
		if err != nil {
			utils.RespondWithJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid request"})
			return
		}

		if errs := validator.Validate(req); errs != nil {
			utils.RespondWithJSON(w, http.StatusBadRequest, map[string]any{"errors": errs})
			return
		}

		l, err := linkService.CreateLink(ctx, userID, req.OriginalURL, req.PrettyID)
		if err != nil {
			if err == link.ErrPrettyIDExists {
				utils.RespondWithJSON(w, http.StatusConflict, map[string]any{"error": "pretty_id already taken"})
				return
			}
			slog.Error("HandleCreateShortLink", "error", err)
			utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to create link"})
			return
		}

		utils.RespondWithJSON(w, http.StatusCreated, map[string]any{"data": newLinkResponse(l)})
	})
}

func HandleListLinks(ctx context.Context, linkService *link.LinkService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(userIDKey).(string)
		if !ok {
			utils.RespondWithJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			return
		}

		links, err := linkService.ListLinks(ctx, userID)
		if err != nil {
			slog.Error("HandleListLinks", "error", err)
			utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to list links"})
			return
		}

		resp := make([]linkResponse, len(links))
		for i, l := range links {
			resp[i] = newLinkResponse(l)
		}

		utils.RespondWithJSON(w, http.StatusOK, map[string]any{"data": resp})
	})
}

func HandleDeleteLink(ctx context.Context, linkService *link.LinkService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(userIDKey).(string)
		if !ok {
			utils.RespondWithJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			return
		}

		linkID := r.PathValue("id")
		if linkID == "" {
			utils.RespondWithJSON(w, http.StatusBadRequest, map[string]any{"error": "missing link id"})
			return
		}

		err := linkService.DeleteLink(ctx, userID, linkID)
		if err != nil {
			slog.Error("HandleDeleteLink", "error", err)
			utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to delete link"})
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, map[string]any{"message": "link deleted"})
	})
}

func HandlePrettyLink(ctx context.Context, validator validation.Validator, linkService *link.LinkService) http.Handler {
	type request struct {
		LinkID   string `json:"link_id" validate:"required"`
		PrettyID string `json:"pretty_id" validate:"required"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(userIDKey).(string)
		if !ok {
			utils.RespondWithJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			return
		}

		req, err := utils.DecodeToJSON[request](r)
		if err != nil {
			utils.RespondWithJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid request"})
			return
		}

		if errs := validator.Validate(req); errs != nil {
			utils.RespondWithJSON(w, http.StatusBadRequest, map[string]any{"errors": errs})
			return
		}

		l, err := linkService.UpdateLinkPrettyID(ctx, userID, req.LinkID, req.PrettyID)
		if err != nil {
			slog.Error("HandlePrettyLink", "error", err)
			utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to update pretty link"})
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, map[string]any{"data": newLinkResponse(l)})
	})
}

func HandleGetLinkAnalytics(ctx context.Context, analyticsService *analytics.AnalyticsService, linkService *link.LinkService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(userIDKey).(string)
		if !ok {
			utils.RespondWithJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			return
		}

		linkID := r.PathValue("id")
		if linkID == "" {
			utils.RespondWithJSON(w, http.StatusBadRequest, map[string]any{"error": "missing link id"})
			return
		}

		l, err := linkService.GetLink(ctx, userID, linkID)
		if err != nil {
			if err == link.ErrLinkNotFound {
				utils.RespondWithJSON(w, http.StatusNotFound, map[string]any{"error": "link not found"})
				return
			}
			slog.Error("HandleGetLinkAnalytics", "error", err)
			utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to get link"})
			return
		}

		if l.UserID != userID {
			utils.RespondWithJSON(w, http.StatusForbidden, map[string]any{"error": "forbidden"})
			return
		}

		stats, err := analyticsService.GetStats(ctx, linkID)
		if err != nil {
			slog.Error("HandleGetLinkAnalytics", "error", err)
			utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to get stats"})
			return
		}

		details, err := analyticsService.GetAnalytics(ctx, linkID)
		if err != nil {
			slog.Error("HandleGetLinkAnalytics", "error", err)
			utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to get details"})
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, map[string]any{
			"stats": stats,
			"details": details,
		})
	})
}
