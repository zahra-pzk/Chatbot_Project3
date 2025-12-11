-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS ai_knowledge (
    id                      BIGSERIAL,
    knowledge_external_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_chunk_id         UUID,
    source_text             TEXT NOT NULL,
    source_meta             JSONB DEFAULT '{}'::jsonb,
    embedding_vector        FLOAT8[],
    embedding_json          JSONB,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by              UUID
);

CREATE INDEX IF NOT EXISTS idx_ai_knowledge_created_at
ON ai_knowledge(created_at);

CREATE INDEX IF NOT EXISTS idx_ai_knowledge_text_gin
ON ai_knowledge USING GIN (to_tsvector('simple', coalesce(source_text,'')));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_ai_knowledge_text_gin;
DROP INDEX IF EXISTS idx_ai_knowledge_created_at;
DROP TABLE IF EXISTS ai_knowledge;
-- +goose StatementEnd
