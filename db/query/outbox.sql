-- name: CreateOutbox :one
INSERT INTO outbox (
    id,
    event_type,
    aggregate_id,
    topic,
    payload,
    created_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
) RETURNING *;

-- name: IncrementRetryOutbox :exec
UPDATE outbox
SET retry_count = retry_count + 1
WHERE id IN (sqlc.arg(ids));

-- name: MarkProcessedOutbox :exec
UPDATE outbox
SET processed_at = sqlc.arg(processed_at)
WHERE id = ANY(sqlc.arg(ids)::uuid[]);

-- name: GetUnprocessedOutbox :many
SELECT * FROM outbox
WHERE retry_count < sqlc.arg(max_retries)
AND processed_at IS NULL
ORDER BY created_at
LIMIT $1;
