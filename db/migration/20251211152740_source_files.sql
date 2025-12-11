-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS source_files (
    source_id           BIGINT,
    source_external_id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    storage_key         TEXT NOT NULL,
    filename            TEXT,
    mime_type           TEXT,
    size_bytes          BIGINT,
    uploaded_by         UUID,
    uploaded_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    processed_at        TIMESTAMP WITH TIME ZONE,
    status              TEXT DEFAULT 'uploaded'
);

CREATE INDEX IF NOT EXISTS idx_source_files_uploaded_at
ON source_files(uploaded_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_source_files_uploaded_at;
DROP TABLE IF EXISTS source_files;
-- +goose StatementEnd
