SELECT
    id,
    account_id,
    subscription_id,
    status,
    currency,
    due_date,
    issued_at,
    paid_at,
    created_at,
    updated_at
FROM invoices
WHERE account_id = :account_id
