SELECT
    id,
    user_id,
    token,
    expires_at,
    created_at,
    used
FROM reset_tokens
WHERE token = :token
