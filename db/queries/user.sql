-- name: CreateUser :exec
INSERT INTO users (id, email, first_name, last_name, password) VALUES (?, ?, ?, ?, ?);

-- name: VerifyUserById :exec
UPDATE users SET email_verified = ?, updated_at = ? WHERE id = ?;

-- name: DeactivateUser :exec
UPDATE users SET active = false WHERE id = ?;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ? LIMIT 1;

-- name: GetUser :one
SELECT * FROM users WHERE id = ? LIMIT 1;
