-- +goose Up
-- +goose StatementBegin
CREATE TABLE sessions (
    session_id          BIGSERIAL,
    session_external_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_agent          VARCHAR(255) NOT NULL,
    username            VARCHAR(255) NOT NULL,
    user_external_id    UUID NOT NULL,
    is_blocked          BOOLEAN NOT NULL DEFAULT false,
    client_ip           VARCHAR(255) NOT NULL,
    refresh_token       TEXT NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at          TIMESTAMPTZ NOT NULL
);

ALTER TABLE sessions
ADD FOREIGN KEY (username) REFERENCES users(username);

ALTER TABLE sessions
ADD FOREIGN KEY (user_external_id) REFERENCES users(user_external_id);

CREATE INDEX idx_sessions_user ON sessions(user_external_id);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);


CREATE OR REPLACE FUNCTION update_user_last_seen()
RETURNS trigger AS $$
BEGIN
    UPDATE users
    SET last_seen = now(), updated_at = now()
    WHERE user_external_id = NEW.user_external_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_last_seen
AFTER INSERT OR UPDATE ON sessions
FOR EACH ROW EXECUTE FUNCTION update_user_last_seen();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_update_last_seen ON sessions;
DROP FUNCTION IF EXISTS update_user_last_seen;
DROP INDEX IF EXISTS idx_sessions_expires;
DROP INDEX IF EXISTS idx_sessions_user;
DROP TABLE IF EXISTS sessions;
-- +goose StatementEnd
