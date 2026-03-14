SELECT
    id,
    payment_id,
    amount_cents,
    reason,
    status,
    created_at,
    updated_at
FROM refunds
WHERE payment_id = :payment_id
