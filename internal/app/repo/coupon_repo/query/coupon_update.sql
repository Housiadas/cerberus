UPDATE
    coupons
SET "code"            = :code,
    "discount_type"   = :discount_type,
    "discount_value"  = :discount_value,
    "currency"        = :currency,
    "max_redemptions" = :max_redemptions,
    "times_redeemed"  = :times_redeemed,
    "is_active"       = :is_active,
    "expires_at"      = :expires_at,
    "updated_at"      = :updated_at
WHERE id = :id