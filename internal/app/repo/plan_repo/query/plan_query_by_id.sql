SELECT
    id,
    name,
    description,
    interval,
    price_cents,
    currency,
    is_active,
    created_at,
    updated_at
FROM plans
WHERE id = :id
