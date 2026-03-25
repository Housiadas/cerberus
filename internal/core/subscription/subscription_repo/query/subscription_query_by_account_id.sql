SELECT
    id,
    account_id,
    stripe_subscription_id,
    stripe_customer_id,
    stripe_price_id,
    status,
    current_period_start,
    current_period_end,
    cancel_at_period_end,
    canceled_at,
    created_at,
    updated_at
FROM subscriptions
WHERE account_id = :account_id
ORDER BY created_at DESC
