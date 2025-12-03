-- Description: Create table users
CREATE TABLE users
(
    id            UUID        NOT NULL,
    name          TEXT        NOT NULL,
    email         TEXT        UNIQUE NOT NULL,
    password_hash TEXT        NOT NULL,
    department    TEXT        NULL,
    enabled       BOOLEAN     NOT NULL,
    date_created  TIMESTAMP   NOT NULL,
    date_updated  TIMESTAMP   NOT NULL,

    PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS users_email_idx ON users ("email");
