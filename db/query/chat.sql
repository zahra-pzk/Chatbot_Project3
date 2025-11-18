-- name: CreateChat :one
INSERT INTO chats (
    user_external_id
) VALUES (
    $1
)
RETURNING *;

-- name: GetChat :one
SELECT * FROM chats
WHERE chat_external_id = $1;

-- name: ListChats :many
SELECT * FROM chats
ORDER BY created_at DESC
LIMIT $1
OFFSET $2;

-- name: UpdateChatStatus :exec
UPDATE chats
SET status = $2,
    updated_at = NOW()
WHERE chat_external_id = $1;

-- name: DeleteChatStatus :exec
DELETE FROM chats
WHERE chat_external_id = $1;