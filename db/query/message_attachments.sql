-- name: CreateAttachment :one
INSERT INTO message_attachments (
    message_external_id,
    user_external_id,
    chat_external_id,
    url,
    filename,
    mime_type,
    size_bytes,
    metadata,
    created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, NOW()
)
RETURNING attachment_id, attachment_external_id, message_external_id, user_external_id, chat_external_id, url, filename, mime_type, size_bytes, metadata, created_at;

-- name: GetAttachment :one
SELECT attachment_id, attachment_external_id, message_external_id, user_external_id, chat_external_id, url, filename, mime_type, size_bytes, metadata, created_at
FROM message_attachments
WHERE attachment_external_id = $1
LIMIT 1;

-- name: ListAttachmentsByMessage :many
SELECT attachment_id, attachment_external_id, message_external_id, user_external_id, chat_external_id, url, filename, mime_type, size_bytes, metadata, created_at
FROM message_attachments
WHERE message_external_id = $1
ORDER BY created_at ASC
LIMIT $2 OFFSET $3;

-- name: ListAllAttachmentsByMessage :many
SELECT attachment_id, attachment_external_id, message_external_id, user_external_id, chat_external_id, url, filename, mime_type, size_bytes, metadata, created_at
FROM message_attachments
WHERE message_external_id = $1
ORDER BY created_at ASC;

-- name: ListAttachmentsByChat :many
SELECT attachment_id, attachment_external_id, message_external_id, user_external_id, chat_external_id, url, filename, mime_type, size_bytes, metadata, created_at
FROM message_attachments
WHERE chat_external_id = $1
ORDER BY created_at ASC
LIMIT $2 OFFSET $3;

-- name: DeleteAttachment :exec
DELETE FROM message_attachments
WHERE attachment_external_id = $1;

-- name: DeleteAttachmentsByMessage :exec
DELETE FROM message_attachments
WHERE message_external_id = $1;

-- name: CountAttachmentsByMessage :one
SELECT COUNT(*) AS count
FROM message_attachments
WHERE message_external_id = $1;

-- name: GetAttachmentsMetadataByChat :many
SELECT message_external_id, COUNT(*) AS attachment_count, SUM(COALESCE(size_bytes,0)) AS total_bytes
FROM message_attachments
WHERE chat_external_id = $1
GROUP BY message_external_id
ORDER BY total_bytes DESC
LIMIT $2 OFFSET $3;
