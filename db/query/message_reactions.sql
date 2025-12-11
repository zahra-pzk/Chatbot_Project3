-- name: AddOrUpdateReaction :one
INSERT INTO message_reactions (
    message_external_id,
    user_external_id,
    reaction,
    score,
    created_at
) VALUES (
    $1, $2, $3, $4, NOW()
)
ON CONFLICT (message_external_id, user_external_id, reaction)
DO UPDATE SET
    score = EXCLUDED.score,
    created_at = NOW()
RETURNING reaction_id, reaction_external_id, message_external_id, user_external_id, reaction, score, created_at;

-- name: RemoveReaction :exec
DELETE FROM message_reactions
WHERE message_reactions.message_external_id = $1 
  AND message_reactions.user_external_id = $2 
  AND message_reactions.reaction = $3;

-- name: ListReactionsByMessage :many
SELECT 
    mr.reaction_id, 
    mr.reaction_external_id, 
    mr.message_external_id, 
    mr.user_external_id, 
    mr.reaction, 
    mr.score, 
    mr.created_at
FROM message_reactions mr
WHERE mr.message_external_id = $1
ORDER BY mr.created_at ASC
LIMIT $2 OFFSET $3;

-- name: ListAllReactionsByMessage :many
SELECT 
    mr.reaction_id, 
    mr.reaction_external_id, 
    mr.message_external_id, 
    mr.user_external_id, 
    mr.reaction, 
    mr.score, 
    mr.created_at
FROM message_reactions mr
WHERE mr.message_external_id = $1
ORDER BY mr.created_at ASC;

-- name: GetUserReactionForMessage :one
SELECT 
    mr.reaction_id, 
    mr.reaction_external_id, 
    mr.message_external_id, 
    mr.user_external_id, 
    mr.reaction, 
    mr.score, 
    mr.created_at
FROM message_reactions mr
WHERE mr.message_external_id = $1 
  AND mr.user_external_id = $2 
  AND mr.reaction = $3
LIMIT 1;

-- name: CountReactionsByMessage :many
SELECT 
    mr.reaction, 
    COUNT(*) AS reaction_count, 
    SUM(mr.score) AS total_score
FROM message_reactions mr
WHERE mr.message_external_id = $1
GROUP BY mr.reaction
ORDER BY total_score DESC;

-- name: CountAllReactionsByMessage :one
SELECT COUNT(*) AS total_count
FROM message_reactions mr
WHERE mr.message_external_id = $1;

-- name: CountUserReactions :one
SELECT COUNT(*) AS user_total_reactions
FROM message_reactions mr
WHERE mr.user_external_id = $1;

-- name: GetReactionsSummaryForChat :many
SELECT 
    mr.reaction, 
    COUNT(*) AS reaction_count, 
    SUM(mr.score) AS total_score
FROM message_reactions mr
JOIN messages m ON m.message_external_id = mr.message_external_id
WHERE m.chat_external_id = $1
GROUP BY mr.reaction
ORDER BY total_score DESC
LIMIT $2 OFFSET $3;

-- name: GetTopReactionersInChat :many
SELECT 
    mr.user_external_id, 
    COUNT(*) AS reactions_count, 
    SUM(mr.score) AS total_score
FROM message_reactions mr
JOIN messages m ON m.message_external_id = mr.message_external_id
WHERE m.chat_external_id = $1
GROUP BY mr.user_external_id
ORDER BY total_score DESC
LIMIT $2 OFFSET $3;

-- name: InsertReactionWithWeight :one
INSERT INTO message_reactions (
    message_external_id,
    user_external_id,
    reaction,
    score,
    created_at
)
VALUES (
    $1, $2, $3,
    COALESCE($4, reaction_weight($3)),
    NOW()
)
ON CONFLICT (message_external_id, user_external_id, reaction)
DO UPDATE SET 
    score = EXCLUDED.score,
    created_at = NOW()
RETURNING 
    reaction_id,
    reaction_external_id,
    message_external_id,
    user_external_id,
    reaction,
    score,
    created_at;

-- name: ToggleReaction :one
WITH target AS (
    DELETE FROM message_reactions
    WHERE message_reactions.message_external_id = $1 
      AND message_reactions.user_external_id = $2 
      AND message_reactions.reaction = $3
    RETURNING 1 AS deleted_val
),
inserted AS (
    INSERT INTO message_reactions (message_external_id, user_external_id, reaction, score, created_at)
    SELECT 
        CAST($1 AS UUID), 
        CAST($2 AS UUID), 
        CAST($3 AS TEXT), 
        COALESCE(CAST($4 AS BIGINT), reaction_weight(CAST($3 AS TEXT))), 
        NOW()
    WHERE NOT EXISTS (SELECT 1 FROM target)
    RETURNING 1 AS inserted_val
)
SELECT 
    EXISTS (SELECT 1 FROM inserted) AS is_inserted,
    COALESCE((
        SELECT SUM(score_mr.score)
        FROM message_reactions score_mr
        WHERE score_mr.message_external_id IN (
            SELECT target_m.message_external_id 
            FROM messages target_m 
            WHERE target_m.chat_external_id = (
                SELECT chat_m.chat_external_id 
                FROM messages chat_m 
                WHERE chat_m.message_external_id = $1 
                LIMIT 1
            )
        )
    ), 0)::BIGINT AS chat_total_score;

-- name: RecomputeChatScore :one
SELECT update_chat_score($1) AS result_score;