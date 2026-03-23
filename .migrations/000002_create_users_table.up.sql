-- Description: Create table users
CREATE TABLE users
(
    id            UUID        NOT NULL,
    account_id    UUID        NULL,
    name          TEXT        NOT NULL,
    email         TEXT UNIQUE NOT NULL,
    password_hash TEXT        NOT NULL,
    department    TEXT        NULL,
    enabled       BOOLEAN     NOT NULL,
    created_at    TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP   NOT NULL,
    deleted_at    TIMESTAMP   NULL     DEFAULT NULL,

    PRIMARY KEY (id),
    FOREIGN KEY (account_id) REFERENCES accounts (id)
);

CREATE INDEX idx_users_deleted_at ON users (id) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_account_id ON users (account_id) WHERE account_id IS NOT NULL;
