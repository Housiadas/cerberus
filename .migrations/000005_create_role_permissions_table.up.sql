-- Description: Create table role_permissions
CREATE TABLE role_permissions
(
    role_id                UUID   NOT NULL,
    permission_id          UUID   NOT NULL,

    FOREIGN KEY (role_id)       REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);
