UPDATE
    payments
SET "status"     = :status,
    "paid_at"    = :paid_at,
    "updated_at" = :updated_at
WHERE id = :id
