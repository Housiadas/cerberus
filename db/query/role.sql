-- name: CreateRole :one
INSERT INTO roles (
    id,
    name,
    created_at,
    updated_at
) VALUES (
    $1,
    $2,
    $3,
    $4
) RETURNING *;

-- name: DeleteRole :exec
UPDATE roles
SET deleted_at = NOW()
WHERE id = sqlc.arg(id)
AND deleted_at IS NULL;

-- name: GetRoleByID :one
SELECT
    id,
    name,
    created_at,
    updated_at
FROM roles
WHERE id = sqlc.arg(id)
AND deleted_at IS NULL;

-- name: UpdateRole :one
UPDATE roles
SET
    name = COALESCE(sqlc.narg(name), name),
    updated_at = COALESCE(sqlc.narg(updated_at), updated_at)
WHERE id = sqlc.arg(id)
RETURNING *;
