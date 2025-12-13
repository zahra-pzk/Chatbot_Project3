-- name: CreateSourceFile :one
INSERT INTO source_files (
  storage_key, filename, mime_type, size_bytes, uploaded_by
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetSourceByExternalID :one
SELECT * FROM source_files
WHERE source_external_id = $1;

-- name: MarkSourceProcessed :exec
UPDATE source_files
SET processed_at = now(),
    status = 'processed'
WHERE source_id = $1;

-- name: ListUploadedSources :many
SELECT *
FROM source_files
ORDER BY uploaded_at DESC
LIMIT $1;

-- name: ListDocumentsByUser :many
SELECT source_id, source_external_id, filename, mime_type, size_bytes, uploaded_at, status
FROM source_files
WHERE uploaded_by = $1
ORDER BY uploaded_at DESC;