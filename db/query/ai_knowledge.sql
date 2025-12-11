-- name: CreateKnowledge :one
INSERT INTO ai_knowledge (
  source_chunk_id, source_text, source_meta, embedding_vector, embedding_json, created_by
) VALUES (
  $1, $2, COALESCE($3, '{}'::jsonb), $4, $5, $6
)
RETURNING *;

-- name: GetKnowledgeByID :one
SELECT * FROM ai_knowledge
WHERE knowledge_external_id = $1
LIMIT 1;

-- name: SearchKnowledgeFulltext :many
SELECT 
    id, knowledge_external_id, source_text, ts_rank_cd(to_tsvector('simple', source_text), to_tsquery('simple', $1)) AS rank
FROM ai_knowledge
WHERE to_tsvector('simple', source_text) @@ to_tsquery('simple', $1)
ORDER BY rank DESC
LIMIT $2;

-- name: UpdateKnowledgeEmbedding :exec
UPDATE ai_knowledge
SET embedding_vector = $2,
    embedding_json = $3
WHERE knowledge_external_id = $1;

-- name: DeleteKnowledge :exec
DELETE FROM ai_knowledge
WHERE knowledge_external_id = $1;
