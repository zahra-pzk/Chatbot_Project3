-- +goose Up
-- +goose StatementBegin
CREATE TABLE messages (
    message_id              BIGSERIAL,
    message_external_id     UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_external_id        UUID            NOT NULL, 
    sender_external_id      UUID            NOT NULL, 
    content                 TEXT            NOT NULL,
    is_system_message       BOOLEAN         NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITHOUT TIME ZONE  NOT NULL,
    updated_at TIMESTAMP WITHOUT TIME ZONE  NOT NULL,
    CONSTRAINT fk_messages_chat
        FOREIGN KEY (chat_external_id)
        REFERENCES chats (chat_external_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_messages_sender
        FOREIGN KEY (sender_external_id)
        REFERENCES users (user_external_id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_messages_chat_id ON messages(chat_external_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_messages_chat_id;
DROP TABLE IF EXISTS messages;
-- +goose StatementEnd
