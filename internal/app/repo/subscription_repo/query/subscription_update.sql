UPDATE subscriptions
SET
    "status"               = :status,
    "current_period_start" = :current_period_start,
    "current_period_end"   = :current_period_end,
    "cancel_at_period_end" = :cancel_at_period_end,
    "canceled_at"          = :canceled_at,
    "updated_at"           = :updated_at
WHERE id = :id
