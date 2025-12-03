-- Description: Create table permissions
CREATE TABLE permissions
(
    id            UUID         NOT NULL,
    name          VARCHAR(100) NOT NULL,
    date_created  TIMESTAMP    NOT NULL,
    date_updated  TIMESTAMP    NOT NULL,

    PRIMARY KEY (id)
);
