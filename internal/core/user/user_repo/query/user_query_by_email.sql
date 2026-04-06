SELECT
    id,
    name,
    email,
    password_hash,
    department,
    enabled,
    account_id,
    created_at,
    updated_at
FROM users
WHERE email = :email
AND deleted_at IS NULL
