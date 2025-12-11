-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE message_type AS ENUM ('user', 'admin', 'system', 'superadmin', 'guest');

ALTER TABLE messages
  ADD COLUMN IF NOT EXISTS message_type message_type NOT NULL DEFAULT 'user';

CREATE TABLE IF NOT EXISTS message_attachments (
    attachment_id           BIGSERIAL,
    attachment_external_id  UUID                        PRIMARY KEY DEFAULT gen_random_uuid(),
    message_external_id     UUID                        NOT NULL,
    user_external_id        UUID                        NOT NULL,
    chat_external_id        UUID                        NOT NULL,
    url                     TEXT                        NOT NULL,
    filename                TEXT,
    mime_type               TEXT,
    size_bytes              BIGINT,
    metadata                JSONB                       DEFAULT '{}'::jsonb,
    created_at              TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_attachments_message FOREIGN KEY (message_external_id)
        REFERENCES messages (message_external_id) ON DELETE CASCADE,
    CONSTRAINT fk_attachments_user
        FOREIGN KEY (user_external_id)
        REFERENCES users (user_external_id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_attachments_chat
        FOREIGN KEY (chat_external_id)
        REFERENCES chats (chat_external_id)
        ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_attachments_message ON message_attachments(message_external_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_attachments_message;
DROP TABLE IF EXISTS message_attachments;
ALTER TABLE messages DROP COLUMN IF EXISTS message_type;
DROP TYPE IF EXISTS message_type;
-- +goose StatementEnd
