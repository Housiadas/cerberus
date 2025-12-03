-- Description: Create table users
CREATE TABLE users
(
    id            UUID        NOT NULL,
    name          TEXT        NOT NULL,
    email         TEXT        UNIQUE NOT NULL,
    role_id       UUID        NOT NULL,
    password_hash TEXT        NOT NULL,
    department    TEXT        NULL,
    enabled       BOOLEAN     NOT NULL,
    date_created  TIMESTAMP   NOT NULL,
    date_updated  TIMESTAMP   NOT NULL,

    FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE,

    PRIMARY KEY (id)
);