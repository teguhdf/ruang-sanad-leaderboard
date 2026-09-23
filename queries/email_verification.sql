-- name: CreateEmailVerification :exec
INSERT INTO email_verifications (token, user_id, email, expires_at, created_at)
VALUES (?, ?, ?, ?, ?);

-- name: GetEmailVerification :one
SELECT * FROM email_verifications WHERE token = ? AND used = 0 AND expires_at > ?;

-- name: MarkEmailVerificationUsed :execrows
UPDATE email_verifications SET used = 1 WHERE token = ?;

-- name: MarkEmailVerified :execrows
UPDATE users SET email_verified = 1, updated_at = ? WHERE id = ?;

-- name: DeleteExpiredEmailVerifications :exec
DELETE FROM email_verifications WHERE expires_at < CURRENT_TIMESTAMP;
