UPDATE refresh_tokens
SET revoked = true
WHERE token = :token
