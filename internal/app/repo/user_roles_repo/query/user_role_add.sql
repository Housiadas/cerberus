INSERT INTO user_roles (user_id, role_id, created_at, updated_at)
VALUES (:user_id, :role_id, :created_at, :updated_at)
ON CONFLICT (user_id, role_id) DO NOTHING