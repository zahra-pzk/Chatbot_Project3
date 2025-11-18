-- name: CreateMessage :one
INSERT INTO messages (
    chat_external_id,
    sender_external_id,
    content,
    is_system_message,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, NOW(), NOW()
)
RETURNING *;

-- name: GetMessage :one
SELECT * FROM messages
WHERE message_external_id = $1;

-- name: ListMessagesByChat :many
SELECT * FROM messages
WHERE chat_external_id = $1
ORDER BY created_at ASC
LIMIT $2
OFFSET $3;

-- name: ListRecentMessagesByChat :many
SELECT * FROM messages
WHERE chat_external_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: UpdateMessage :one
UPDATE messages
SET content = $2,
    updated_at = NOW()
WHERE message_external_id = $1
RETURNING *;

-- name: DeleteMessage :exec
DELETE FROM messages
WHERE message_external_id = $1;
