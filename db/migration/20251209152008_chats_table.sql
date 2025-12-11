-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE TYPE chat_status_type AS ENUM ('open', 'pending', 'closed');

CREATE TABLE chats (
    chat_id             BIGSERIAL,
    chat_external_id    UUID                PRIMARY KEY DEFAULT gen_random_uuid(),
    user_external_id    UUID                NOT NULL,
    label               VARCHAR(255)        NOT NULL,
    status              chat_status_type    NOT NULL DEFAULT 'pending',
    admin_external_id   UUID,
    score               BIGINT,
    created_at  TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    updated_at  TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    CONSTRAINT fk_chats_user
        FOREIGN KEY (user_external_id) 
        REFERENCES users (user_external_id) 
        ON DELETE RESTRICT,
    CONSTRAINT fk_chats_admin
        FOREIGN KEY (admin_external_id) 
        REFERENCES users (user_external_id) 
        ON DELETE RESTRICT
);

CREATE INDEX idx_chats_status ON chats(status);
CREATE INDEX idx_chats_score ON chats(score);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_chats_score;
DROP INDEX IF EXISTS idx_chats_status;
DROP TABLE IF EXISTS chats;
DROP TYPE IF EXISTS chat_status_type;
-- +goose StatementEnd
