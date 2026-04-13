-- name: GetPermissionsByUserID :many
SELECT DISTINCT permission_id, permission_name
FROM vw_user_roles_permissions
WHERE user_id = sqlc.arg(user_id)
AND permission_name IS NOT NULL;
