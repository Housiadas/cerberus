-- Description: Create table roles
CREATE TABLE roles
(
    id            UUID         NOT NULL,
    name          VARCHAR(100) UNIQUE NOT NULL,
    created_at    TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP   NOT NULL,
    deleted_at    TIMESTAMP   NULL DEFAULT NULL,

    PRIMARY KEY (id)
);

CREATE INDEX idx_roles_deleted_at ON roles (id) WHERE deleted_at IS NULL;
