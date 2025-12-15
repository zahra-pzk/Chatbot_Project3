-- name: CreateChat :one
INSERT INTO chats (
    user_external_id,
    status,
    label,
    admin_external_id,
    score,
    created_at,
    updated_at
) VALUES (
    $1, $2::chat_status_type, $3, $4, $5, NOW(), NOW()
)
RETURNING chat_id, chat_external_id, user_external_id, label, status, admin_external_id, score, created_at, updated_at;

-- name: CreateChatDefaults :one
INSERT INTO chats (
    user_external_id,
    label,
    created_at,
    updated_at
) VALUES (
    $1, $2, NOW(), NOW()
)
RETURNING chat_id, chat_external_id, user_external_id, label, status, admin_external_id, score, created_at, updated_at;

-- name: AssignedAdminToChat :one
UPDATE chats
SET admin_external_id = $2,
    updated_at = NOW()
WHERE chat_external_id = $1
RETURNING chat_id, chat_external_id, user_external_id, label, status, admin_external_id, score, created_at, updated_at;

-- name: GetChat :one
SELECT chat_id, chat_external_id, user_external_id, label, status, admin_external_id, score, created_at, updated_at
FROM chats
WHERE chat_external_id = $1
LIMIT 1;

-- name: GetChatsByUser :many
SELECT chat_id, chat_external_id, user_external_id, label, status, admin_external_id, score, created_at, updated_at
FROM chats
WHERE user_external_id = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET $3;

-- name: ListChats :many
SELECT chat_id, chat_external_id, user_external_id, label, status, admin_external_id, score, created_at, updated_at
FROM chats
ORDER BY updated_at DESC
LIMIT $1
OFFSET $2;

-- name: UpdateChatStatus :one
UPDATE chats
SET status = $2::chat_status_type,
    updated_at = NOW()
WHERE chat_external_id = $1
RETURNING chat_id, chat_external_id, user_external_id, label, status, admin_external_id, score, created_at, updated_at;

-- name: UpdateChat :one
UPDATE chats
SET
    user_external_id = COALESCE(NULLIF($2, '00000000-0000-0000-0000-000000000000'::uuid), user_external_id),
    status = COALESCE($3::chat_status_type, status),
    label = COALESCE(NULLIF($4, ''), label),
    admin_external_id = COALESCE(NULLIF($5, '00000000-0000-0000-0000-000000000000'::uuid), admin_external_id),
    updated_at = NOW()
WHERE chat_external_id = $1
RETURNING chat_id, chat_external_id, user_external_id, label, status, admin_external_id, score, created_at, updated_at;

-- name: UpdateChatScore :one
UPDATE chats
SET score = $2,
    updated_at = NOW()
WHERE chat_external_id = $1
RETURNING chat_id, chat_external_id, user_external_id, label, status, admin_external_id, score, created_at, updated_at;

-- name: DeleteChat :exec
DELETE FROM chats
WHERE chat_external_id = $1;

-- name: GetOpenChatByUser :one
SELECT chat_id, chat_external_id, user_external_id, label, status, admin_external_id, score, created_at, updated_at
FROM chats
WHERE user_external_id = $1
  AND status = 'open'::chat_status_type
ORDER BY created_at DESC
LIMIT 1
FOR UPDATE;

-- name: GetPendingChatByUser :one
SELECT chat_id, chat_external_id, user_external_id, label, status, admin_external_id, score, created_at, updated_at
FROM chats
WHERE user_external_id = $1
  AND status = 'pending'::chat_status_type
ORDER BY created_at DESC
LIMIT 1
FOR UPDATE;

-- name: GetClosedChatByUser :one
SELECT chat_id, chat_external_id, user_external_id, label, status, admin_external_id, score, created_at, updated_at
FROM chats
WHERE user_external_id = $1
  AND status = 'closed'::chat_status_type
ORDER BY created_at DESC
LIMIT 1
FOR UPDATE;

-- name: ListPendingChats :many
SELECT chat_id, chat_external_id, user_external_id, label, status, admin_external_id, score, created_at, updated_at
FROM chats
WHERE status = 'pending'::chat_status_type
ORDER BY updated_at DESC
LIMIT $1
OFFSET $2;

-- name: ListOpenChats :many
SELECT chat_id, chat_external_id, user_external_id, label, status, admin_external_id, score, created_at, updated_at
FROM chats
WHERE status = 'open'::chat_status_type
ORDER BY updated_at DESC
LIMIT $1
OFFSET $2;

-- name: ListClosedChats :many
SELECT chat_id, chat_external_id, user_external_id, label, status, admin_external_id, score, created_at, updated_at
FROM chats
WHERE status = 'closed'::chat_status_type
ORDER BY updated_at DESC
LIMIT $1
OFFSET $2;

-- name: GetChatsByAdmin :many
SELECT chat_id, chat_external_id, user_external_id, label, status, admin_external_id, score, created_at, updated_at
FROM chats
WHERE admin_external_id = $1
ORDER BY updated_at DESC
LIMIT $2
OFFSET $3;

-- name: CountUserChats :one
SELECT COUNT(*) AS count
FROM chats
WHERE user_external_id = $1;

-- name: GetChatsByStatusAndScoreRange :many
SELECT chat_id, chat_external_id, user_external_id, label, status, admin_external_id, score, created_at, updated_at
FROM chats
WHERE status = $1
  AND ($2 IS NULL OR score >= $2)
  AND ($3 IS NULL OR score <= $3)
ORDER BY updated_at DESC
LIMIT $4
OFFSET $5;

-- name: GetTopChatsByScore :many
SELECT chat_id, chat_external_id, user_external_id, label, status, admin_external_id, score, created_at, updated_at
FROM chats
ORDER BY score DESC NULLS LAST, updated_at DESC
LIMIT $1
OFFSET $2;
