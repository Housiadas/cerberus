-- Description: Create RBAC tables (roles, permissions, role_permissions, user_roles)
CREATE TABLE roles
(
    id         UUID         NOT NULL,
    name       VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP    NOT NULL,
    deleted_at TIMESTAMP    NULL     DEFAULT NULL,

    PRIMARY KEY (id)
);

CREATE INDEX idx_roles_deleted_at ON roles (id) WHERE deleted_at IS NULL;

CREATE TABLE permissions
(
    id         UUID         NOT NULL,
    name       VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP    NOT NULL,
    deleted_at TIMESTAMP    NULL     DEFAULT NULL,

    PRIMARY KEY (id)
);

CREATE INDEX idx_permissions_deleted_at ON permissions (id) WHERE deleted_at IS NULL;

CREATE TABLE role_permissions
(
    role_id       UUID      NOT NULL,
    permission_id UUID      NOT NULL,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP NOT NULL,
    deleted_at    TIMESTAMP NULL     DEFAULT NULL,

    FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions (id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE user_roles
(
    user_id    UUID      NOT NULL,
    role_id    UUID      NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP NULL     DEFAULT NULL,

    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);
