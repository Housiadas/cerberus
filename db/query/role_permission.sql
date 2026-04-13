-- name: CreateRolePermission :one
INSERT INTO role_permissions (
    role_id,
    permission_id,
    created_at,
    updated_at
) VALUES (
    $1,
    $2,
    $3,
    $4
) ON CONFLICT (role_id, permission_id) DO NOTHING RETURNING *;

-- name: DeleteRolePermission :exec
DELETE FROM role_permissions
WHERE role_id = sqlc.arg(role_id)
AND permission_id = sqlc.arg(permission_id);
