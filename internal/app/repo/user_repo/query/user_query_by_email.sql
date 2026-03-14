SELECT
    id,
    name,
    email,
    password_hash,
    department,
    account_id,
    enabled,
    created_at,
    updated_at
FROM users
WHERE email = :email
AND deleted_at IS NULL
