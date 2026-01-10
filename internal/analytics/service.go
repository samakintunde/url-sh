package analytics

import (
	"context"
	"database/sql"
	"log/slog"
	db "url-shortener/db/sqlc"
	"url-shortener/internal/utils"
)

type AnalyticsService struct {
	queries *db.Queries
}

func NewAnalyticsService(queries *db.Queries) *AnalyticsService {
	return &AnalyticsService{
		queries: queries,
	}
}

type ClickStats struct {
	TotalClicks  int64 `json:"total_clicks"`
	UniqueClicks int64 `json:"unique_clicks"`
}

type AnalyticsData struct {
	ID          string `json:"id"`
	LinkID      string `json:"link_id"`
	IpAddress   string `json:"ip_address,omitempty"`
	UserAgent   string `json:"user_agent,omitempty"`
	CountryCode string `json:"country_code,omitempty"`
	CreatedAt   string `json:"created_at"`
}

func (s *AnalyticsService) RecordClick(ctx context.Context, linkID, ip, userAgent string) error {
	id := utils.NewULID().String()

	// Basic country code logic could be added here if we had an IP geolocation service.
	// For now, we leave it null or implement a placeholder if needed.
	countryCode := sql.NullString{}

	_, err := s.queries.CreateLinkClick(ctx, db.CreateLinkClickParams{
		ID:          id,
		LinkID:      linkID,
		IpAddress:   sql.NullString{String: ip, Valid: ip != ""},
		UserAgent:   sql.NullString{String: userAgent, Valid: userAgent != ""},
		CountryCode: countryCode,
	})

	if err != nil {
		slog.Error("service.analytics.RecordClick", "error", err)
		return err
	}
	return nil
}

func (s *AnalyticsService) GetAnalytics(ctx context.Context, linkID string) ([]AnalyticsData, error) {
	data, err := s.queries.GetLinkAnalytics(ctx, linkID)
	if err != nil {
		slog.Error("service.analytics.GetAnalytics", "error", err)
		return nil, err
	}

	result := make([]AnalyticsData, len(data))
	for i, d := range data {
		result[i] = AnalyticsData{
			ID:          d.ID,
			LinkID:      d.LinkID,
			IpAddress:   d.IpAddress.String,
			UserAgent:   d.UserAgent.String,
			CountryCode: d.CountryCode.String,
			CreatedAt:   utils.ConvertTimeToString(d.CreatedAt),
		}
	}
	return result, nil
}

func (s *AnalyticsService) GetStats(ctx context.Context, linkID string) (ClickStats, error) {
	stats, err := s.queries.GetLinkAnalyticsStats(ctx, linkID)
	if err != nil {
		return ClickStats{}, err
	}
	return ClickStats{
		TotalClicks:  stats.TotalClicks,
		UniqueClicks: stats.UniqueClicks,
	}, nil
}
