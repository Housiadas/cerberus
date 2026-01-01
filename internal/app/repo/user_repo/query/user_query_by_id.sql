SELECT
    id,
    role_id,
    name,
    email,
    password_hash,
    department,
    enabled,
    created_at,
    updated_at
FROM users
WHERE user_id = :user_id