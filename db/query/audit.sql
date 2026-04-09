-- name: CreateAudit :one
INSERT INTO audit (
    id,
    obj_id,
    obj_entity,
    obj_name,
    actor_id,
    action,
    data,
    message,
    created_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9
) RETURNING *;
