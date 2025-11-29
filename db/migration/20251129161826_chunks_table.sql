-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS chunks (
        chunk_external_id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        id SERIAL,
        text TEXT NOT NULL,
        embedding JSONB NOT NULL
    );

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS chunks;
-- +goose StatementEnd