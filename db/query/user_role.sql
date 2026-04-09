-- name: CreateUserRole :one
INSERT INTO user_roles (
    user_id,
    role_id,
    created_at,
    updated_at
) VALUES (
    $1,
    $2,
    $3,
    $4
) ON CONFLICT (user_id, role_id) DO NOTHING RETURNING *;

-- name: DeleteUserRole :exec
DELETE FROM user_roles
WHERE user_id = sqlc.arg(user_id)
AND role_id = sqlc.arg(role_id);
