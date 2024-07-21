-- name: CreateUser :one
INSERT INTO users (id, email, first_name, last_name, password) VALUES (?, ?, ?, ?, ?) RETURNING *;

-- name: VerifyUserById :exec
UPDATE users SET email_verified = ?, status = "active" WHERE id = ?;

-- name: DeactivateUser :exec
UPDATE users SET status = "inactive" WHERE id = ?;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ? LIMIT 1;

-- name: DoesUserExistByEmail :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = ?);

-- name: GetUser :one
SELECT * FROM users WHERE id = ? LIMIT 1;

-- name: DoesUserExist :one
SELECT EXISTS(SELECT 1 FROM users WHERE id = ?);
