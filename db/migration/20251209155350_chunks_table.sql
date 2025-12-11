-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS chunks (
    chunk_internal_id   BIGSERIAL,
    chunk_external_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id           UUID,
    source_path         TEXT,
    source_filename     TEXT,
    source_mime         TEXT,
    source_page         INT,
    department          TEXT,
    language            TEXT,
    text                TEXT NOT NULL,
    text_tsv            tsvector,
    embedding_vector    BYTEA,
    embedding_json      JSONB NOT NULL,
    chunk_hash          TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by          UUID,
    status              TEXT DEFAULT 'ready'
);

CREATE OR REPLACE FUNCTION chunks_tsv_trigger()
RETURNS trigger AS $$
BEGIN
    NEW.text_tsv := to_tsvector('simple', coalesce(NEW.text,''));
    RETURN NEW;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

CREATE TRIGGER trg_chunks_tsv
BEFORE INSERT OR UPDATE ON chunks
FOR EACH ROW EXECUTE FUNCTION chunks_tsv_trigger();

CREATE INDEX IF NOT EXISTS idx_chunks_department ON chunks(department);
CREATE INDEX IF NOT EXISTS idx_chunks_tsv ON chunks USING GIN (text_tsv);
CREATE INDEX IF NOT EXISTS idx_chunks_hash ON chunks(chunk_hash);
CREATE INDEX IF NOT EXISTS idx_chunks_created_at ON chunks(created_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_chunks_tsv ON chunks;
DROP FUNCTION IF EXISTS chunks_tsv_trigger;
DROP INDEX IF EXISTS idx_chunks_created_at;
DROP INDEX IF EXISTS idx_chunks_hash;
DROP INDEX IF EXISTS idx_chunks_tsv;
DROP INDEX IF EXISTS idx_chunks_department;
DROP TABLE IF EXISTS chunks;
-- +goose StatementEnd