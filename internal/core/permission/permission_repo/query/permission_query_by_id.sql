SELECT
    id,
    name,
    created_at,
    updated_at
FROM permissions
WHERE id = :id
AND deleted_at IS NULL
