-- name: CreateUserAccount :one
INSERT INTO user_account (
    user_external_id,
    updated_at
) VALUES (
    $1, NOW()
)
RETURNING *;

-- name: GetUserAccountByExternalID :one
SELECT
    u.name,
    u.username,
    u.email,
    u.role,
    a.account_external_id,
    a.status,
    a.birth_date,
    a.photos
FROM users u
INNER JOIN user_account a 
    ON u.user_external_id = a.user_external_id
WHERE u.user_external_id = $1
LIMIT 1;

-- name: GetUserAccountByAccountID :one
SELECT
    u.name,
    u.username,
    u.email,
    u.role,
    a.account_external_id,
    a.status,
    a.birth_date,
    a.photos
FROM user_account a
INNER JOIN users u
    ON a.user_external_id = u.user_external_id
WHERE a.account_external_id = $1
LIMIT 1;

-- name: DeleteUserAccount :exec
DELETE FROM user_account
WHERE account_external_id = $1;

-- name: UpdateUserAccountProfile :one
UPDATE user_account
SET
    birth_date = $2,
    photos = $3,
    updated_at = NOW()
WHERE user_external_id = $1
RETURNING *;

-- name: UpdateUserAccountStatus :one
UPDATE user_account
SET
    status = $2,
    updated_at = NOW()
WHERE user_external_id = $1
RETURNING *;
