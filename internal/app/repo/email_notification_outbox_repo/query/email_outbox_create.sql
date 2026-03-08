INSERT INTO email_notification_outbox
(
    id,
    event_type,
    to_email,
    payload,
    created_at
)
VALUES
(
    :id,
    :event_type,
    :to_email,
    :payload,
    :created_at
)
