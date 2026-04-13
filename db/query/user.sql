-- name: CreateUser :one
INSERT INTO users (
    id,
    account_id,
    name,
    email,
    password_hash,
    department,
    enabled,
    created_at,
    updated_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9
) RETURNING *;

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = sqlc.arg(id)
AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = sqlc.arg(email)
AND deleted_at IS NULL;

-- name: UpdateUser :one
UPDATE users
SET
    name = COALESCE(sqlc.narg(name), name),
    email = COALESCE(sqlc.narg(email), email),
    password_hash = COALESCE(sqlc.narg(password_hash), password_hash),
    department = COALESCE(sqlc.narg(department), department),
    enabled = COALESCE(sqlc.narg(enabled), enabled),
    updated_at = COALESCE(sqlc.narg(updated_at), updated_at)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteUser :exec
UPDATE users
SET deleted_at = NOW()
WHERE id = sqlc.arg(id)
AND deleted_at IS NULL;
