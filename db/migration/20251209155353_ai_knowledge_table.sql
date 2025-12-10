-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS ai_knowledge (
    id BIGSERIAL,
    knowledge_external_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chunk TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    embedding JSONB
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS ai_knowledge;
-- +goose StatementEnd
