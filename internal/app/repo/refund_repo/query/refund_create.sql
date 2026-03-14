INSERT INTO refunds
(id, payment_id, amount_cents, reason, status, created_at, updated_at)
VALUES (:id, :payment_id, :amount_cents, :reason, :status, :created_at, :updated_at)
