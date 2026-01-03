-- Description: Create table users
CREATE TABLE users
(
    id            UUID        NOT NULL,
    name          TEXT        NOT NULL,
    email         TEXT        UNIQUE NOT NULL,
    password_hash TEXT        NOT NULL,
    department    TEXT        NULL,
    enabled       BOOLEAN     NOT NULL,
    created_at    TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP   NOT NULL,
    last_login    TIMESTAMP,

    PRIMARY KEY (id)
);
