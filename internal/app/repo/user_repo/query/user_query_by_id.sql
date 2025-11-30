SELECT
    id,
    role_id,
    name,
    email,
    password_hash,
    department,
    enabled,
    date_created,
    date_updated
FROM users
WHERE user_id = :user_id