-- Description: Create outbox tables (outbox, email_notification_outbox)
CREATE TABLE outbox
(
    id           UUID         NOT NULL,
    event_type   VARCHAR(100) NOT NULL,
    aggregate_id UUID         NOT NULL,
    topic        VARCHAR(255) NOT NULL,
    payload      JSONB        NOT NULL,
    retry_count  INT          NOT NULL DEFAULT 0,
    created_at   TIMESTAMP    NOT NULL,
    processed_at TIMESTAMP    NULL     DEFAULT NULL,

    PRIMARY KEY (id)
);

CREATE INDEX idx_outbox_unprocessed
    ON outbox (created_at)
    WHERE processed_at IS NULL;

-- Description: Add partial indexes on processed_at for outbox cleanup queries
CREATE INDEX idx_outbox_processed
    ON outbox (processed_at)
    WHERE processed_at IS NOT NULL;

CREATE TABLE email_notification_outbox
(
    id           UUID         NOT NULL,
    event_type   VARCHAR(100) NOT NULL,
    to_email     VARCHAR(255) NOT NULL,
    payload      JSONB        NOT NULL,
    retry_count  INT          NOT NULL DEFAULT 0,
    created_at   TIMESTAMP    NOT NULL,
    processed_at TIMESTAMP    NULL     DEFAULT NULL,

    PRIMARY KEY (id)
);

CREATE INDEX idx_email_notification_outbox_unprocessed
    ON email_notification_outbox (created_at)
    WHERE processed_at IS NULL;

CREATE INDEX idx_email_notification_outbox_processed
    ON email_notification_outbox (processed_at)
    WHERE processed_at IS NOT NULL;
