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
    :id,
    :user_id,
    :token,
    :expires_at,
    :created_at,
    :revoked
)
