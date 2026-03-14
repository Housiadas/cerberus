SELECT
    id,
    invoice_id,
    payment_method_id,
    amount_cents,
    currency,
    status,
    paid_at,
    created_at,
    updated_at
FROM payments
WHERE id = :id
