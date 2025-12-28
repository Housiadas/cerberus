-- Description: Create table roles
CREATE TABLE roles
(
    id            UUID         NOT NULL,
    name          VARCHAR(100) UNIQUE NOT NULL,
    date_created  TIMESTAMP    NOT NULL,
    date_updated  TIMESTAMP    NOT NULL,

    PRIMARY KEY (id)
);
