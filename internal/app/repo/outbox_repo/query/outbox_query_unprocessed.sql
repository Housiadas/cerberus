SELECT
    id,
    event_type,
    aggregate_id,
    topic,
    payload,
    created_at,
    processed_at
FROM
    outbox
WHERE
    processed_at IS NULL
ORDER BY
    created_at
LIMIT
    :limit
