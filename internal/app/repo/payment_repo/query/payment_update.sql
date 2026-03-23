UPDATE payments
SET
    "status"     = :status,
    "updated_at" = :updated_at
WHERE id = :id
