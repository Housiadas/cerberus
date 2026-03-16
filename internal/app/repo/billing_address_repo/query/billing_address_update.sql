UPDATE billing_addresses
SET
    "line1"       = :line1,
    "line2"       = :line2,
    "city"        = :city,
    "state"       = :state,
    "postal_code" = :postal_code,
    "country"     = :country,
    "updated_at"  = :updated_at
WHERE id = :id
