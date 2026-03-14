UPDATE tax_rates
SET "name"        = :name,
    "percentage"  = :percentage,
    "country"     = :country,
    "description" = :description,
    "is_active"   = :is_active,
    "updated_at"  = :updated_at
WHERE id = :id
