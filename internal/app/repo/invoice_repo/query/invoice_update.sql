UPDATE invoices
SET
    "status"      = :status,
    "amount_paid" = :amount_paid,
    "updated_at"  = :updated_at
WHERE id = :id
