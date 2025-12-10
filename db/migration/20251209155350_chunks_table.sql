-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS chunks (
    chunk_external_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    id SERIAL,
    text TEXT NOT NULL,
    embedding JSONB NOT NULL,
    department TEXT
);
CREATE INDEX IF NOT EXISTS idx_chunks_department ON chunks(department);
CREATE INDEX IF NOT EXISTS idx_chunks_text_gin ON chunks USING gin (to_tsvector('simple', text));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS chunks;
-- +goose StatementEnd