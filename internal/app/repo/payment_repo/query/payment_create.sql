INSERT INTO payments
(id, invoice_id, payment_method_id, amount_cents, currency, status, paid_at, created_at, updated_at)
VALUES (:id, :invoice_id, :payment_method_id, :amount_cents, :currency, :status, :paid_at, :created_at, :updated_at)
