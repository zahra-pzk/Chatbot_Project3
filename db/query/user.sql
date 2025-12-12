-- name: CreateUser :one
INSERT INTO users (
    first_name,
    last_name,
    username,
    phone_number,
    email,
    hashed_password,
    role,
    status,
    birth_date,
    photos,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW()
)
RETURNING user_id, user_external_id, first_name, last_name, username, phone_number, email, hashed_password, role, created_at, updated_at, status, birth_date, photos;

-- name: CreateGuestUser :one
INSERT INTO users (
    first_name,
    last_name,
    email,
    role,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, NOW(), NOW()
)
RETURNING user_id, user_external_id, first_name, last_name, username, phone_number, email, hashed_password, role, created_at, updated_at, status, birth_date, photos;

-- name: GetUser :one
SELECT user_id, user_external_id, first_name, last_name, username, phone_number, email, hashed_password, role, created_at, updated_at, status, birth_date, photos, last_seen
FROM users
WHERE user_external_id = $1
LIMIT 1;

-- name: GetUserByEmail :one
SELECT user_id, user_external_id, first_name, last_name, username, phone_number, email, hashed_password, role, created_at, updated_at, status, birth_date, photos, last_seen
FROM users
WHERE email = $1
LIMIT 1;

-- name: GetUserByUsername :one
SELECT user_id, user_external_id, first_name, last_name, username, phone_number, email, hashed_password, role, created_at, updated_at, status, birth_date, photos, last_seen
FROM users
WHERE username = $1
LIMIT 1;

-- name: GetUserByPhoneNumber :one
SELECT user_id, user_external_id, first_name, last_name, username, phone_number, email, hashed_password, role, created_at, updated_at, status, birth_date, photos, last_seen
FROM users
WHERE phone_number = $1
LIMIT 1;

-- name: ListUsers :many
SELECT user_id, user_external_id, first_name, last_name, username, phone_number, email, hashed_password, role, created_at, updated_at, status, birth_date, photos, last_seen
FROM users
ORDER BY created_at DESC
LIMIT $1
OFFSET $2;

-- name: UpdateUser :one
UPDATE users
SET
    first_name = COALESCE(NULLIF($2, ''), first_name),
    last_name = COALESCE(NULLIF($3, ''), last_name),
    username = COALESCE(NULLIF($4, ''), username),
    phone_number = COALESCE(NULLIF($5, ''), phone_number),
    email = COALESCE(NULLIF($6, ''), email),
    role = COALESCE($7, role),
    status = COALESCE($8, status),
    birth_date = COALESCE($9, birth_date),
    photos = COALESCE($10, photos),
    updated_at = NOW()
WHERE user_external_id = $1
RETURNING user_id, user_external_id, first_name, last_name, username, phone_number, email, hashed_password, role, created_at, updated_at, status, birth_date, photos, last_seen;

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
  first_name,
  last_name, 
  username, 
  phone_number, 
  email, 
  hashed_password,
  role, 
  created_at, 
  updated_at,
  status, 
  birth_date, 
  photos,
  last_seen
FROM users 
WHERE user_external_id = $1
LIMIT 1;

-- name: UpdateUserStatus :one
UPDATE users
SET
    status = $2,
    updated_at = NOW()
WHERE user_external_id = $1
RETURNING user_id, user_external_id, first_name, last_name, username, phone_number, email, hashed_password, role, created_at, updated_at, status, birth_date, photos, last_seen;

-- name: AddPhotoToUserProfile :one
UPDATE users
SET  
    photos = array_append(photos, $2),
    updated_at = NOW()
WHERE user_external_id = $1
RETURNING user_id, user_external_id, first_name, last_name, username, phone_number, email, hashed_password, role, created_at, updated_at, status, birth_date, photos, last_seen;

-- name: AddPhotosToUserProfile :one
UPDATE users
SET 
    photos = photos || $2,
    updated_at = NOW()
WHERE user_external_id = $1
RETURNING user_id, user_external_id, first_name, last_name, username, phone_number, email, hashed_password, role, created_at, updated_at, status, birth_date, photos, last_seen;

-- name: UpdateUserRole :one
UPDATE users
SET
    role = $2,
    updated_at = NOW()
WHERE user_external_id = $1
RETURNING user_id, user_external_id, first_name, last_name, username, phone_number, email, hashed_password, role, created_at, updated_at, status, birth_date, photos, last_seen;
