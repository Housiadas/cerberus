-- Description: Create table permissions
CREATE TABLE permissions
(
    id            UUID         NOT NULL,
    name          VARCHAR(100) UNIQUE NOT NULL,
    created_at    TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP   NOT NULL,

    PRIMARY KEY (id)
);
