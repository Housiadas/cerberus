SELECT
    id,
    account_id,
    payment_id,
    stripe_refund_id,
    amount,
    currency,
    status,
    reason,
    created_at,
    updated_at
FROM refunds
WHERE account_id = :account_id
ORDER BY created_at DESC
