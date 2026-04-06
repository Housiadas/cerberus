SELECT
    id,
    account_id,
    stripe_invoice_id,
    stripe_customer_id,
    status,
    amount_due,
    amount_paid,
    currency,
    invoice_url,
    period_start,
    period_end,
    created_at,
    updated_at
FROM invoices
WHERE account_id = :account_id
ORDER BY created_at DESC
