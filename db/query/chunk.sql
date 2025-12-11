-- name: CreateChunk :one
INSERT INTO chunks (
  source_id, source_path, source_filename, source_mime, source_page, department, language, text, embedding_vector, embedding_json, chunk_hash, created_by, status
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, COALESCE($13, 'ready')
)
RETURNING *;

-- name: GetChunkByID :one
SELECT * FROM chunks
WHERE chunk_external_id = $1 LIMIT 1;

-- name: ListChunksBySource :many
SELECT * FROM chunks
WHERE source_id = $1
ORDER BY created_at DESC;

-- name: SearchChunksFulltext :many
SELECT chunk_external_id, text, ts_rank_cd(text_tsv, query) AS rank
FROM chunks, to_tsquery('simple', $1) AS query
WHERE text_tsv @@ query
ORDER BY rank DESC
LIMIT $2;

-- name: UpdateChunkStatus :exec
UPDATE chunks
SET status = $2, created_at = created_at
WHERE chunk_external_id = $1;

-- name: UpdateChunkEmbedding :exec
UPDATE chunks
SET embedding_vector = $2,
    embedding_json = $3
WHERE chunk_external_id = $1;

-- name: DeleteChunk :exec
DELETE FROM chunks
WHERE chunk_external_id = $1;
