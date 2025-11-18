-- name: CreateChat :one
INSERT INTO chats (
    user_external_id,
    status,
    created_at,
    updated_at
) VALUES (
    $1, $2, NOW(), NOW()
)
RETURNING *;

-- name: CreateChatDefaults :one
INSERT INTO chats (
    user_external_id,
    created_at,
    updated_at
) VALUES (
    $1, NOW(), NOW()
)
RETURNING *;

-- name: GetChat :one
SELECT * FROM chats
WHERE chat_external_id = $1;

-- name: GetChatsByUser :many
SELECT * FROM chats
WHERE user_external_id = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET $3;

-- name: ListChats :many
SELECT * FROM chats
ORDER BY created_at DESC
LIMIT $1
OFFSET $2;

-- name: UpdateChatStatus :one
UPDATE chats
SET status = $2,
    updated_at = NOW()
WHERE chat_external_id = $1
RETURNING *;

-- name: UpdateChat :one
UPDATE chats
SET
    user_external_id = COALESCE($2, user_external_id),
    status = COALESCE($3, status),
    updated_at = NOW()
WHERE chat_external_id = $1
RETURNING *;

-- name: DeleteChat :exec
DELETE FROM chats
WHERE chat_external_id = $1;
