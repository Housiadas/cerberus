UPDATE
    subscriptions
SET "status"               = :status,
    "current_period_start" = :current_period_start,
    "current_period_end"   = :current_period_end,
    "updated_at"           = :updated_at
WHERE id = :id
