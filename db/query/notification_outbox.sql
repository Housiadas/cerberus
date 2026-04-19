-- name: CreateNotificationOutbox :one
INSERT INTO email_notification_outbox (
    id,
    event_type,
    to_email,
    payload,
    created_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
) RETURNING *;

-- name: IncrementRetryNotificationOutbox :exec
UPDATE email_notification_outbox
SET retry_count = retry_count + 1
WHERE id = ANY(sqlc.arg(ids)::uuid[]);

-- name: MarkProcessedNotificationOutbox :exec
UPDATE email_notification_outbox
SET processed_at = sqlc.arg(processed_at)
WHERE id = ANY(sqlc.arg(ids)::uuid[]);

-- name: GetUnprocessedNotificationOutbox :many
SELECT * FROM email_notification_outbox
WHERE
retry_count < sqlc.arg(max_retries)
AND processed_at IS NULL
ORDER BY created_at
LIMIT $1;

-- name: DeleteProcessedNotificationOutbox :execrows
DELETE FROM email_notification_outbox
WHERE processed_at IS NOT NULL
  AND processed_at < @before::timestamp;
