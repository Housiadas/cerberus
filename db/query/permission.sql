-- name: CreatePermission :one
INSERT INTO permissions (
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

-- name: DeletePermission :exec
UPDATE permissions
SET deleted_at = NOW()
WHERE id = sqlc.arg(id)
AND deleted_at IS NULL;

-- name: GetPermissionByID :one
SELECT
    id,
    name,
    created_at,
    updated_at
FROM permissions
WHERE id = sqlc.arg(id)
AND deleted_at IS NULL;

-- name: UpdatePermission :one
UPDATE permissions
SET
    name = COALESCE(sqlc.narg(name), name),
    updated_at = COALESCE(sqlc.narg(updated_at), updated_at)
WHERE id = sqlc.arg(id)
RETURNING *;
