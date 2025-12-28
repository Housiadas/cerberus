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
FROM refresh_tokens
WHERE user_id = :user_id
