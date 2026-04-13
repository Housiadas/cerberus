-- name: CreateAccount :one
INSERT INTO accounts (
    id,
    name,
    stripe_customer_id,
    created_at,
    updated_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
) RETURNING *;

-- name: DeleteAccount :exec
UPDATE accounts
SET deleted_at = NOW()
WHERE id = sqlc.arg(id)
AND deleted_at IS NULL;

-- name: GetAccountByID :one
SELECT
    id,
    name,
    stripe_customer_id,
    created_at,
    updated_at
FROM accounts
WHERE id = sqlc.arg(id)
AND deleted_at IS NULL;

-- name: UpdateAccount :one
UPDATE accounts
SET
    name = COALESCE(sqlc.narg(name), name),
    stripe_customer_id = COALESCE(sqlc.narg(stripe_customer_id), stripe_customer_id),
    updated_at = COALESCE(sqlc.narg(updated_at), updated_at)
WHERE id = sqlc.arg(id)
RETURNING *;
