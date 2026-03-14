SELECT
    id,
    subscription_id,
    description,
    quantity,
    recorded_at,
    created_at
FROM usage_records
WHERE id = :id
