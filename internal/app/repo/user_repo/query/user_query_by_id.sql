SELECT
    id,
    name,
    email,
    password_hash,
    department,
    enabled,
    created_at,
    updated_at
FROM users
WHERE id = :id
