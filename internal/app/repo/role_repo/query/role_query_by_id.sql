SELECT
    id,
    name,
    created_at,
    updated_at
FROM roles
WHERE id = :id
AND deleted_at IS NULL
