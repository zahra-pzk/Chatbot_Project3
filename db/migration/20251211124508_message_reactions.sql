-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS message_reactions (
    reaction_id          BIGSERIAL,
    reaction_external_id UUID                        PRIMARY KEY DEFAULT gen_random_uuid(),
    message_external_id  UUID                        NOT NULL,
    user_external_id     UUID                        NOT NULL,
    reaction             TEXT                        NOT NULL,
    score                BIGINT                      NOT NULL,
    created_at           TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_reactions_message FOREIGN KEY (message_external_id)
        REFERENCES messages (message_external_id) ON DELETE CASCADE,
    CONSTRAINT fk_reactions_user FOREIGN KEY (user_external_id)
        REFERENCES users (user_external_id) ON DELETE CASCADE,
    CONSTRAINT uniq_reaction_per_user_per_message UNIQUE (message_external_id, user_external_id, reaction)
);

CREATE INDEX IF NOT EXISTS idx_reactions_message ON message_reactions(message_external_id);
CREATE INDEX IF NOT EXISTS idx_reactions_user ON message_reactions(user_external_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_reactions_user;
DROP INDEX IF EXISTS idx_reactions_message;
DROP TABLE IF EXISTS message_reactions;
-- +goose StatementEnd