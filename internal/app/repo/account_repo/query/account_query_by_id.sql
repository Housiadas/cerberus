SELECT
    id,
    name,
    type,
    enabled,
    created_at,
    updated_at
FROM accounts
WHERE id = :id
AND deleted_at IS NULL
