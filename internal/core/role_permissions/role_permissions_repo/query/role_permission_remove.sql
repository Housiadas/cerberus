DELETE FROM role_permissions
WHERE role_id = :role_id
AND permission_id = :permission_id
