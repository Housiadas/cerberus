SELECT
    id,
    user_id,
    token,
    expires_at,
    created_at,
    revoked
FROM refresh_tokens
WHERE token = :token
