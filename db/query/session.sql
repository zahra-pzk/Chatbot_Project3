-- name: CreateSession :one
INSERT INTO sessions (
  user_agent, username, user_external_id, is_blocked, client_ip, refresh_token, expires_at
) VALUES (
  $1, $2, $3, COALESCE($4, false), $5, $6, $7
)
RETURNING *;

-- name: GetSessionByExternalID :one
SELECT * FROM sessions
WHERE session_external_id = $1 LIMIT 1;

-- name: ListSessionsByUser :many
SELECT * FROM sessions
WHERE user_external_id = $1
ORDER BY created_at DESC;

-- name: UpdateSessionToken :one
UPDATE sessions
SET refresh_token = $2,
    updated_at = now(),
    expires_at = $3
WHERE session_external_id = $1
RETURNING *;

-- name: BlockSession :exec
UPDATE sessions
SET is_blocked = true, updated_at = now()
WHERE session_external_id = $1;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE session_external_id = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at < now();

-- name: IsUserOnline :one
SELECT EXISTS (
    SELECT 1 FROM sessions
    WHERE user_external_id = $1
      AND is_blocked = false
      AND expires_at > now()
) AS is_online;

-- name: GetSessionByRefreshToken :one
SELECT session_id, session_external_id, user_agent, username, user_external_id, is_blocked, client_ip, refresh_token, created_at, updated_at, expires_at
FROM sessions
WHERE refresh_token = $1
LIMIT 1;