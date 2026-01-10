package server

import (
	"context"
	"log/slog"
	"net/http"
	"url-shortener/internal/analytics"
	"url-shortener/internal/link"
)

func HandleRedirect(ctx context.Context, linkService *link.LinkService, analyticsService *analytics.AnalyticsService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := r.PathValue("code")
		if code == "" {
			http.NotFound(w, r)
			return
		}

		l, err := linkService.GetLinkToRedirect(ctx, code)
		if err != nil {
			if err == link.ErrLinkNotFound {
				http.NotFound(w, r)
				return
			}
			slog.Error("HandleRedirect", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Record analytics asynchronously
		go func() {
			ip := r.RemoteAddr
			// Handle X-Forwarded-For if behind proxy (often expected in production)
			if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
				ip = fwd
			}
			userAgent := r.UserAgent()

			// New context for async operation
			bgCtx := context.Background()
			if err := analyticsService.RecordClick(bgCtx, l.ID, ip, userAgent); err != nil {
				slog.Error("HandleRedirect", "message", "failed to record click", "error", err)
			}
		}()

		http.Redirect(w, r, l.OriginalUrl, http.StatusFound)
	})
}
