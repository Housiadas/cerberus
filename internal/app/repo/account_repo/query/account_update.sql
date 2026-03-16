UPDATE
    accounts
SET
    "name"               = :name,
    "stripe_customer_id" = :stripe_customer_id,
    "updated_at"         = :updated_at
WHERE id = :id
