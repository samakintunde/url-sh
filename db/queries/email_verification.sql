-- name: CreateEmailVerification :one
INSERT INTO email_verifications (
    user_id,
    email,
    code,
    expires_at
) VALUES (
    ?, ?, ?, ?
) RETURNING *;

-- name: GetEmailVerification :one
SELECT * FROM email_verifications WHERE user_id = ? AND email = ? AND verified_at IS NULL;

-- name: GetEmailVerificationByCode :one
SELECT * FROM email_verifications WHERE code = ? AND verified_at IS NULL;

-- name: UpdateEmailVerification :one
UPDATE email_verifications
SET code = ? AND verified_at = ? AND expires_at = ?
WHERE id = ?
RETURNING *;

-- name: DeleteEmailVerification :exec
DELETE FROM email_verifications
WHERE id = ?;

-- name: CompleteEmailVerification :one
UPDATE email_verifications
SET verified_at = CURRENT_TIMESTAMP
WHERE user_id = ? AND email = ? AND code = ? AND expires_at > CURRENT_TIMESTAMP
RETURNING *;

-- name: IsEmailVerificationComplete :one
SELECT EXISTS(
    SELECT 1 FROM email_verifications
    WHERE user_id = ? AND email = ? AND verified_at IS NOT NULL
) AS is_verified;

-- name: CleanExpiredEmailVerifications :exec
DELETE FROM email_verifications
WHERE expires_at < CURRENT_TIMESTAMP AND verified_at IS NULL;

-- name: CleanExpiredEmailVerificationsForUserID :exec
DELETE FROM email_verifications
WHERE user_id = ? AND expires_at < CURRENT_TIMESTAMP AND verified_at IS NULL;

-- name: GetUserUnverifiedEmailVerifications :many
SELECT * FROM email_verifications
WHERE user_id = ? AND verified_at IS NULL;
