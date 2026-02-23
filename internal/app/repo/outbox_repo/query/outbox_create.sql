INSERT INTO outbox
(
    id,
    event_type,
    aggregate_id,
    topic,
    payload,
    created_at
)
VALUES
(
    :id,
    :event_type,
    :aggregate_id,
    :topic,
    :payload,
    :created_at
)
