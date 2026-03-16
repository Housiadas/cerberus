SELECT
    id,
    account_id,
    line1,
    line2,
    city,
    state,
    postal_code,
    country,
    created_at,
    updated_at
FROM billing_addresses
WHERE id = :id
