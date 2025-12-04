-- Description: Create table user_roles
CREATE TABLE user_roles
(
    user_id          UUID   NOT NULL,
    role_id          UUID   NOT NULL,
    date_created     TIMESTAMP   NOT NULL,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);
