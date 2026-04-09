-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens
(
    id,
    user_id,
    token,
    expires_at,
    created_at,
    revoked
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

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked = true
WHERE token = sqlc.arg(token);

-- name: GetRefreshTokenByToken :one
SELECT
    id,
    user_id,
    token,
    expires_at,
    created_at,
    revoked
FROM refresh_tokens
WHERE token = sqlc.arg(token);
