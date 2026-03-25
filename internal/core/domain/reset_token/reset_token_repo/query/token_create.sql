INSERT INTO reset_tokens
(
    id,
    user_id,
    token,
    expires_at,
    created_at,
    used
)
VALUES
(
    :id,
    :user_id,
    :token,
    :expires_at,
    :created_at,
    :used
)
