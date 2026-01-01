CREATE TABLE refresh_tokens (
    id          UUID            NOT NULL,
    user_id     UUID            NOT NULL,
    token       VARCHAR(255)    UNIQUE NOT NULL,
    expires_at  TIMESTAMP       NOT NULL,
    created_at  TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked     BOOLEAN         NOT NULL DEFAULT FALSE

    PRIMARY KEY (id)
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX refresh_tokens_token_idx   ON refresh_tokens(token);
CREATE INDEX refresh_tokens_user_id_idx ON refresh_tokens(user_id);
