INSERT INTO invoices
(id, account_id, subscription_id, status, currency, due_date, issued_at, paid_at, created_at, updated_at)
VALUES (:id, :account_id, :subscription_id, :status, :currency, :due_date, :issued_at, :paid_at, :created_at, :updated_at)
