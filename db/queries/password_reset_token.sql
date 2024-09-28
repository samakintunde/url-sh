-- name: CreatePasswordResetToken :one
INSERT INTO password_reset_tokens (user_id, token, expires_at)
VALUES (?, ?, ?) RETURNING *;

-- name: GetPasswordResetToken :one
SELECT * FROM password_reset_tokens
WHERE user_id = ? AND token = ? AND expires_at > CURRENT_TIMESTAMP;

-- name: GetPasswordResetTokenByUserID :one
SELECT * FROM password_reset_tokens
WHERE user_id = ? AND expires_at > CURRENT_TIMESTAMP;

-- name: DeletePasswordResetToken :exec
DELETE FROM password_reset_tokens WHERE id = ?;
