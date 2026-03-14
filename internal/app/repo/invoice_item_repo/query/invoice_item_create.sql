INSERT INTO invoice_items
(id, invoice_id, description, quantity, unit_price_cents, tax_rate_id, created_at)
VALUES (:id, :invoice_id, :description, :quantity, :unit_price_cents, :tax_rate_id, :created_at)
