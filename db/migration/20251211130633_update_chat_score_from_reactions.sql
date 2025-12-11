-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION reaction_weight(reaction TEXT) RETURNS BIGINT
LANGUAGE SQL IMMUTABLE AS $$
  SELECT CASE reaction
    WHEN '👍' THEN 1
    WHEN '👎' THEN -1
    WHEN '👌' THEN 3
    WHEN '🚩' THEN -5
    ELSE 1
  END;
$$;

CREATE OR REPLACE FUNCTION update_chat_score(p_chat_uuid UUID) RETURNS VOID
LANGUAGE plpgsql AS $$
DECLARE
  v_sum BIGINT;
BEGIN
  SELECT COALESCE(SUM(mr.score), 0) INTO v_sum
  FROM message_reactions mr
  JOIN messages m ON m.message_external_id = mr.message_external_id
  WHERE m.chat_external_id = p_chat_uuid;

  UPDATE chats
  SET score = v_sum,
      updated_at = NOW()
  WHERE chat_external_id = p_chat_uuid;
END;
$$;

CREATE OR REPLACE FUNCTION message_reactions_change_trigger() RETURNS TRIGGER
LANGUAGE plpgsql AS $$
DECLARE
  affected_chat UUID;
  old_chat UUID;
BEGIN
  IF (TG_OP = 'INSERT') THEN
    SELECT m.chat_external_id INTO affected_chat FROM messages m WHERE m.message_external_id = NEW.message_external_id;
    IF affected_chat IS NOT NULL THEN
      PERFORM update_chat_score(affected_chat);
    END IF;
    RETURN NEW;
  ELSIF (TG_OP = 'UPDATE') THEN
    SELECT m.chat_external_id INTO affected_chat FROM messages m WHERE m.message_external_id = NEW.message_external_id;
    SELECT m.chat_external_id INTO old_chat FROM messages m WHERE m.message_external_id = OLD.message_external_id;
    IF old_chat IS NOT NULL AND old_chat <> affected_chat THEN
      PERFORM update_chat_score(old_chat);
    END IF;
    IF affected_chat IS NOT NULL THEN
      PERFORM update_chat_score(affected_chat);
    END IF;
    RETURN NEW;
  ELSIF (TG_OP = 'DELETE') THEN
    SELECT m.chat_external_id INTO old_chat FROM messages m WHERE m.message_external_id = OLD.message_external_id;
    IF old_chat IS NOT NULL THEN
      PERFORM update_chat_score(old_chat);
    END IF;
    RETURN OLD;
  END IF;
  RETURN NULL;
END;
$$;

DROP TRIGGER IF EXISTS trg_message_reactions_change ON message_reactions;
CREATE TRIGGER trg_message_reactions_change
AFTER INSERT OR UPDATE OR DELETE ON message_reactions
FOR EACH ROW EXECUTE FUNCTION message_reactions_change_trigger();

CREATE OR REPLACE FUNCTION recompute_all_chat_scores() RETURNS VOID
LANGUAGE plpgsql AS $$
DECLARE
  r RECORD;
  v_sum BIGINT;
BEGIN
  FOR r IN SELECT chat_external_id FROM chats LOOP
    SELECT COALESCE(SUM(mr.score), 0) INTO v_sum
    FROM message_reactions mr
    JOIN messages m ON m.message_external_id = mr.message_external_id
    WHERE m.chat_external_id = r.chat_external_id;

    UPDATE chats
    SET score = v_sum,
        updated_at = NOW()
    WHERE chat_external_id = r.chat_external_id;
  END LOOP;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_message_reactions_change ON message_reactions;
DROP FUNCTION IF EXISTS message_reactions_change_trigger();
DROP FUNCTION IF EXISTS update_chat_score(uuid);
DROP FUNCTION IF EXISTS reaction_weight(text);
DROP FUNCTION IF EXISTS recompute_all_chat_scores();
-- +goose StatementEnd
