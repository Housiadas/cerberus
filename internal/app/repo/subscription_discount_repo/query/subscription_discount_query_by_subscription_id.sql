SELECT
    id,
    subscription_id,
    coupon_id,
    applied_at,
    created_at
FROM subscription_discounts
WHERE subscription_id = :subscription_id
