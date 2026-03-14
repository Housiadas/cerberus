SELECT
    id,
    invoice_id,
    description,
    quantity,
    unit_price_cents,
    tax_rate_id,
    total_cents,
    created_at
FROM invoice_items
WHERE id = :id
