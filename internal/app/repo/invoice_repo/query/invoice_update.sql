UPDATE
    invoices
SET "status"     = :status,
    "due_date"   = :due_date,
    "issued_at"  = :issued_at,
    "paid_at"    = :paid_at,
    "updated_at" = :updated_at
WHERE id = :id
