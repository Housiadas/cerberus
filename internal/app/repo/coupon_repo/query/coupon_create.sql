INSERT INTO coupons
(id, code, discount_type, discount_value, currency, max_redemptions, times_redeemed, is_active, expires_at, created_at, updated_at)
VALUES (:id, :code, :discount_type, :discount_value, :currency, :max_redemptions, :times_redeemed, :is_active, :expires_at, :created_at, :updated_at)
