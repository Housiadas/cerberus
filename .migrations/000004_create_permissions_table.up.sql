-- Description: Create table permissions
CREATE TABLE permissions
(
    id            UUID         NOT NULL,
    name          VARCHAR(100) UNIQUE NOT NULL,
    created_at    TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP   NOT NULL,
    deleted_at    TIMESTAMP   NULL DEFAULT NULL,

    PRIMARY KEY (id)
);

CREATE INDEX idx_permissions_deleted_at ON permissions (id) WHERE deleted_at IS NULL;
