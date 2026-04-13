-- name: CreateResetToken :one
INSERT INTO reset_tokens
(
    id,
    user_id,
    token,
    used,
    expires_at,
    created_at
)
VALUES
(
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
) RETURNING *;

-- name: DeleteResetToken :exec
DELETE
FROM reset_tokens
WHERE id = sqlc.arg(id);

-- name: GetResetTokenByToken :one
SELECT
    id,
    user_id,
    token,
    expires_at,
    created_at,
    used
FROM reset_tokens
WHERE token = sqlc.arg(token);
