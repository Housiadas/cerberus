-- Description: Create table audit
CREATE TABLE audit
(
    id         UUID      NOT NULL,
    obj_id     UUID      NOT NULL,
    obj_entity TEXT      NOT NULL,
    obj_name   TEXT      NOT NULL,
    actor_id   UUID      NOT NULL,
    action     TEXT      NOT NULL,
    data       JSONB     NULL,
    message    TEXT      NULL,
    created_at  TIMESTAMP NOT NULL,

    PRIMARY KEY (id)
);
