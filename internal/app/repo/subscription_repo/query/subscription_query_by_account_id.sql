SELECT
    id,
    account_id,
    plan_id,
    status,
    current_period_start,
    current_period_end,
    created_at,
    updated_at
FROM subscriptions
WHERE account_id = :account_id
