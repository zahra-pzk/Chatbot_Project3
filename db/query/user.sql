-- name: CreateUser :one
INSERT INTO users (
    name,
    username,
    phone_number,
    email,
    hashed_password,
    role,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, NOW(), NOW()
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE user_external_id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT $1
OFFSET $2;

-- name: UpdateUser :one
UPDATE users
SET
    name = $2,
    username = $3,
    phone_number = $4,
    email = $5,
    role = $6,
    updated_at = NOW()
WHERE user_external_id = $1
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users
SET hashed_password = $2,
    updated_at = NOW()
WHERE user_external_id = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE user_external_id = $1;

-- name: GetUserByExternalID :one
SELECT 
  user_id, 
  user_external_id, 
  name, 
  username, 
  phone_number, 
  email, 
  hashed_password, 
  role, 
  created_at, 
  updated_at 
FROM users 
WHERE user_external_id = $1 LIMIT 1;