-- name: CreateUser :exec
INSERT INTO users (id, email, password) VALUES (?, ?, ?);

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ? LIMIT 1;
