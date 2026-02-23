SELECT
    id,
    event_type,
    aggregate_id,
    topic,
    payload,
    retry_count,
    created_at,
    processed_at
FROM
    outbox
WHERE
    processed_at IS NULL
    AND retry_count < :max_retries
ORDER BY
    created_at
LIMIT
    :limit
