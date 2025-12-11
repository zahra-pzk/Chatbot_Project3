-- name: CreateMessage :one
INSERT INTO messages (
    chat_external_id,
    sender_external_id,
    content,
    is_system_message,
    is_admin_message,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, NOW(), NOW()
)
RETURNING message_id, message_external_id, chat_external_id, sender_external_id, content, is_system_message, is_admin_message, created_at, updated_at;

-- name: GetMessage :one
SELECT message_id, message_external_id, chat_external_id, sender_external_id, content, is_system_message, is_admin_message, created_at, updated_at
FROM messages
WHERE message_external_id = $1
LIMIT 1;

-- name: ListMessagesByChat :many
SELECT message_id, message_external_id, chat_external_id, sender_external_id, content, is_system_message, is_admin_message, created_at, updated_at
FROM messages
WHERE chat_external_id = $1
ORDER BY created_at ASC
LIMIT $2
OFFSET $3;

-- name: ListRecentMessagesByChat :many
SELECT message_id, message_external_id, chat_external_id, sender_external_id, content, is_system_message, is_admin_message, created_at, updated_at
FROM messages
WHERE chat_external_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: EditMessage :one
UPDATE messages
SET content = $2,
    updated_at = NOW()
WHERE message_external_id = $1
RETURNING message_id, message_external_id, chat_external_id, sender_external_id, content, is_system_message, is_admin_message, created_at, updated_at;

-- name: DeleteMessage :exec
DELETE FROM messages
WHERE message_external_id = $1;

-- name: GetLastMessageByChat :one
SELECT message_id, message_external_id, chat_external_id, sender_external_id, content, is_system_message, is_admin_message, created_at, updated_at
FROM messages
WHERE chat_external_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: CountMessagesByChat :one
SELECT COUNT(*) AS count
FROM messages
WHERE chat_external_id = $1;

-- name: ListMessagesByChatSince :many
SELECT message_id, message_external_id, chat_external_id, sender_external_id, content, is_system_message, is_admin_message, created_at, updated_at
FROM messages
WHERE chat_external_id = $1
  AND created_at >= $2
ORDER BY created_at ASC
LIMIT $3
OFFSET $4;

-- name: DeleteMessagesByChat :exec
DELETE FROM messages
WHERE chat_external_id = $1;

-- name: MarkMessageAsAdmin :one
UPDATE messages
SET is_admin_message = TRUE,
    updated_at = NOW()
WHERE message_external_id = $1
RETURNING message_id, message_external_id, chat_external_id, sender_external_id, content, is_system_message, is_admin_message, created_at, updated_at;

-- name: MarkMessageAsSystem :one
UPDATE messages
SET is_system_message = TRUE,
    updated_at = NOW()
WHERE message_external_id = $1
RETURNING message_id, message_external_id, chat_external_id, sender_external_id, content, is_system_message, is_admin_message, created_at, updated_at;
