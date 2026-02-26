SELECT DISTINCT permission_name
FROM vw_user_roles_permissions
WHERE user_id = :user_id AND permission_name IS NOT NULL
