// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.21.0
// source: user.sql

package db

import (
	"context"
	"database/sql"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (id, email, first_name, last_name, password) VALUES (?, ?, ?, ?, ?) RETURNING id, email, first_name, last_name, password, status, last_login_at, created_at, updated_at
`

type CreateUserParams struct {
	ID        string
	Email     string
	FirstName sql.NullString
	LastName  sql.NullString
	Password  string
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, createUser,
		arg.ID,
		arg.Email,
		arg.FirstName,
		arg.LastName,
		arg.Password,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.FirstName,
		&i.LastName,
		&i.Password,
		&i.Status,
		&i.LastLoginAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deactivateUser = `-- name: DeactivateUser :exec
UPDATE users SET status = "inactive" WHERE id = ?
`

func (q *Queries) DeactivateUser(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, deactivateUser, id)
	return err
}

const deleteUser = `-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?
`

func (q *Queries) DeleteUser(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, deleteUser, id)
	return err
}

const doesUserExist = `-- name: DoesUserExist :one
SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)
`

func (q *Queries) DoesUserExist(ctx context.Context, id string) (int64, error) {
	row := q.db.QueryRowContext(ctx, doesUserExist, id)
	var column_1 int64
	err := row.Scan(&column_1)
	return column_1, err
}

const doesUserExistByEmail = `-- name: DoesUserExistByEmail :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)
`

func (q *Queries) DoesUserExistByEmail(ctx context.Context, email string) (int64, error) {
	row := q.db.QueryRowContext(ctx, doesUserExistByEmail, email)
	var column_1 int64
	err := row.Scan(&column_1)
	return column_1, err
}

const getUser = `-- name: GetUser :one
SELECT id, email, first_name, last_name, password, status, last_login_at, created_at, updated_at FROM users WHERE id = ? LIMIT 1
`

func (q *Queries) GetUser(ctx context.Context, id string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUser, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.FirstName,
		&i.LastName,
		&i.Password,
		&i.Status,
		&i.LastLoginAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, email, first_name, last_name, password, status, last_login_at, created_at, updated_at FROM users WHERE email = ? LIMIT 1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByEmail, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.FirstName,
		&i.LastName,
		&i.Password,
		&i.Status,
		&i.LastLoginAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateUserLoginTime = `-- name: UpdateUserLoginTime :exec
UPDATE users SET last_login_at = ? WHERE id = ?
`

type UpdateUserLoginTimeParams struct {
	LastLoginAt sql.NullTime
	ID          string
}

func (q *Queries) UpdateUserLoginTime(ctx context.Context, arg UpdateUserLoginTimeParams) error {
	_, err := q.db.ExecContext(ctx, updateUserLoginTime, arg.LastLoginAt, arg.ID)
	return err
}

const updateUserPassword = `-- name: UpdateUserPassword :exec
UPDATE users SET password = ? WHERE id = ?
`

type UpdateUserPasswordParams struct {
	Password string
	ID       string
}

func (q *Queries) UpdateUserPassword(ctx context.Context, arg UpdateUserPasswordParams) error {
	_, err := q.db.ExecContext(ctx, updateUserPassword, arg.Password, arg.ID)
	return err
}

const verifyUserById = `-- name: VerifyUserById :exec
UPDATE users SET status = "active" WHERE id = ?
`

func (q *Queries) VerifyUserById(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, verifyUserById, id)
	return err
}
