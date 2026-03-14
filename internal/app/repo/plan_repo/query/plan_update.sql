UPDATE plans
SET "name"        = :name,
    "description" = :description,
    "interval"    = :interval,
    "price_cents" = :price_cents,
    "currency"    = :currency,
    "is_active"   = :is_active,
    "updated_at"  = :updated_at
WHERE id = :id
