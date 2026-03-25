INSERT INTO role_permissions (role_id, permission_id, created_at, updated_at)
VALUES (:role_id, :permission_id, :created_at, :updated_at)
ON CONFLICT (role_id, permission_id) DO NOTHING
