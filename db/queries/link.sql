-- name: CreateShortLink :one
INSERT INTO links (id, user_id, original_url, short_url_id, pretty_id) VALUES (?, ?, ?, ?, ?) RETURNING *;

-- name: PrettifyShortLink :one
UPDATE links SET pretty_id = ? WHERE id = ? AND user_id = ? RETURNING *;

-- name: DeleteShortLink :exec
DELETE FROM links WHERE user_id = ? AND id = ?;

-- name: GetShortLinks :many
SELECT * FROM links WHERE user_id = ?;

-- name: GetShortLinkById :one
SELECT * FROM links WHERE user_id = ? AND id = ? LIMIT 1;

-- name: GetShortLinkByShortUrlId :one
SELECT * FROM links WHERE user_id = ? AND short_url_id = ? LIMIT 1;
