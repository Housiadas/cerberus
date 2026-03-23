SELECT
    id,
    account_id,
    stripe_payment_id,
    stripe_invoice_id,
    stripe_customer_id,
    amount,
    currency,
    status,
    created_at,
    updated_at
FROM payments
WHERE stripe_payment_id = :stripe_payment_id
