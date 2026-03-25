INSERT INTO accounts
(
    id,
    name,
    stripe_customer_id,
    created_at,
    updated_at
)
VALUES
(
    :id,
    :name,
    :stripe_customer_id,
    :created_at,
    :updated_at
)
