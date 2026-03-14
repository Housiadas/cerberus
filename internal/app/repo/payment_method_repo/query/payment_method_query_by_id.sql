SELECT
    id,
    account_id,
    type,
    details,
    is_default,
    created_at,
    updated_at
FROM payment_methods
WHERE id = :id
