-- name: CreateEmailVerification :exec
INSERT INTO email_verifications (
    user_id,
    email,
    code,
    expires_at
) VALUES (
    ?, ?, ?, ?
) RETURNING *;

-- name: GetEmailVerification :one
SELECT id, code, expires_at, verified_at FROM email_verifications WHERE user_id = ? AND email = ?;

-- name: CompleteEmailVerification :one
UPDATE email_verifications
SET verified_at = CURRENT_TIMESTAMP
WHERE user_id = ? AND email = ? AND code = ? AND expires_at > CURRENT_TIMESTAMP
RETURNING *;

-- name: CleanExpiredEmailVerifications :exec
DELETE FROM email_verifications
WHERE expires_at < CURRENT_TIMESTAMP OR verified_at IS NOT NULL;

-- name: GetUserUnverifiedEmailVerifications :many
SELECT email, code, created_at, expires_at
FROM email_verifications
WHERE user_id = ? AND verified_at IS NULL;

-- name: IsUserEmailVerificationComplete :one
SELECT EXISTS(
    SELECT 1 FROM email_verifications
    WHERE user_id = ? AND email = ? AND verified_at IS NOT NULL
) AS is_verified;

-- name: RecreateEmailVerification :exec
UPDATE email_verifications
SET code = ?, expires_at = ?, created_at = CURRENT_TIMESTAMP
WHERE user_id = ? AND email = ? AND verified_at IS NULL;
