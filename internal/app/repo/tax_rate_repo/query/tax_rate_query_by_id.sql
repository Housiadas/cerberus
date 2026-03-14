SELECT
    id,
    name,
    percentage,
    country,
    description,
    is_active,
    created_at,
    updated_at
FROM tax_rates
WHERE id = :id
