CREATE TABLE email_notification_outbox (
    id           UUID         NOT NULL,
    event_type   VARCHAR(100) NOT NULL,
    to_email     VARCHAR(255) NOT NULL,
    payload      JSONB        NOT NULL,
    retry_count  INT          NOT NULL DEFAULT 0,
    created_at   TIMESTAMP    NOT NULL,
    processed_at TIMESTAMP    NULL DEFAULT NULL,

    PRIMARY KEY (id)
);

CREATE INDEX idx_email_notification_outbox_unprocessed
    ON email_notification_outbox (created_at)
    WHERE processed_at IS NULL;
