SELECT
    id,
    name,
    stripe_customer_id,
    created_at,
    updated_at
FROM accounts
WHERE id = :id
AND deleted_at IS NULL
