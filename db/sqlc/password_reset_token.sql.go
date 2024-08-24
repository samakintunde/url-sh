// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.21.0
// source: password_reset_token.sql

package db

import (
	"context"
	"time"
)

const createPasswordResetToken = `-- name: CreatePasswordResetToken :one
INSERT INTO password_reset_tokens (user_id, token, expires_at)
VALUES (?, ?, ?) RETURNING id, user_id, token, expires_at, created_at
`

type CreatePasswordResetTokenParams struct {
	UserID    string
	Token     string
	ExpiresAt time.Time
}

func (q *Queries) CreatePasswordResetToken(ctx context.Context, arg CreatePasswordResetTokenParams) (PasswordResetToken, error) {
	row := q.db.QueryRowContext(ctx, createPasswordResetToken, arg.UserID, arg.Token, arg.ExpiresAt)
	var i PasswordResetToken
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Token,
		&i.ExpiresAt,
		&i.CreatedAt,
	)
	return i, err
}

const deletePasswordResetToken = `-- name: DeletePasswordResetToken :exec
DELETE FROM password_reset_tokens WHERE token = ?
`

func (q *Queries) DeletePasswordResetToken(ctx context.Context, token string) error {
	_, err := q.db.ExecContext(ctx, deletePasswordResetToken, token)
	return err
}

const getPasswordResetToken = `-- name: GetPasswordResetToken :one
SELECT id, user_id, token, expires_at, created_at FROM password_reset_tokens
WHERE token = ? AND expires_at > datetime('now')
`

func (q *Queries) GetPasswordResetToken(ctx context.Context, token string) (PasswordResetToken, error) {
	row := q.db.QueryRowContext(ctx, getPasswordResetToken, token)
	var i PasswordResetToken
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Token,
		&i.ExpiresAt,
		&i.CreatedAt,
	)
	return i, err
}

const getPasswordResetTokenByUserID = `-- name: GetPasswordResetTokenByUserID :one
SELECT id, user_id, token, expires_at, created_at FROM password_reset_tokens
WHERE user_id = ? AND expires_at > datetime('now')
`

func (q *Queries) GetPasswordResetTokenByUserID(ctx context.Context, userID string) (PasswordResetToken, error) {
	row := q.db.QueryRowContext(ctx, getPasswordResetTokenByUserID, userID)
	var i PasswordResetToken
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Token,
		&i.ExpiresAt,
		&i.CreatedAt,
	)
	return i, err
}
