-- name: CreateLinkClick :one
INSERT INTO link_analytics (id, link_id, ip_address, user_agent, country_code) VALUES (?, ?, ?, ?, ?) RETURNING *;

-- name: GetLinkAnalytics :many
SELECT * FROM link_analytics WHERE link_id = ? ORDER BY created_at DESC;

-- name: GetLinkAnalyticsStats :one
SELECT COUNT(*) as total_clicks, COUNT(DISTINCT ip_address) as unique_clicks FROM link_analytics WHERE link_id = ?;
