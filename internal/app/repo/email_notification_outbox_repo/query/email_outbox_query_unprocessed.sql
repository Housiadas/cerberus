SELECT
    id,
    event_type,
    to_email,
    payload,
    retry_count,
    created_at,
    processed_at
FROM
    email_notification_outbox
WHERE
    processed_at IS NULL
    AND retry_count < :max_retries
ORDER BY
    created_at
LIMIT
    :limit
