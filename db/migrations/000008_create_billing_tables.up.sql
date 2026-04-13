-- Description: Create billing tables (subscriptions, invoices, payments, refunds)
CREATE TABLE subscriptions
(
    id                     UUID         NOT NULL,
    account_id             UUID         NOT NULL REFERENCES accounts (id),
    stripe_subscription_id VARCHAR(255) NOT NULL UNIQUE,
    stripe_customer_id     VARCHAR(255) NOT NULL,
    stripe_price_id        VARCHAR(255) NOT NULL,
    status                 VARCHAR(50)  NOT NULL,
    current_period_start   TIMESTAMP    NULL,
    current_period_end     TIMESTAMP    NULL,
    cancel_at_period_end   BOOLEAN      NOT NULL DEFAULT FALSE,
    canceled_at            TIMESTAMP    NULL,
    created_at             TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at             TIMESTAMP    NOT NULL,

    PRIMARY KEY (id)
);

CREATE INDEX idx_subscriptions_account_id ON subscriptions (account_id);
CREATE INDEX idx_subscriptions_stripe_subscription_id ON subscriptions (stripe_subscription_id);

CREATE TABLE invoices
(
    id                 UUID         NOT NULL,
    account_id         UUID         NOT NULL REFERENCES accounts (id),
    stripe_invoice_id  VARCHAR(255) NOT NULL UNIQUE,
    stripe_customer_id VARCHAR(255) NOT NULL,
    status             VARCHAR(50)  NOT NULL,
    amount_due         BIGINT       NOT NULL DEFAULT 0,
    amount_paid        BIGINT       NOT NULL DEFAULT 0,
    currency           VARCHAR(10)  NOT NULL DEFAULT 'usd',
    invoice_url        TEXT         NULL,
    period_start       TIMESTAMP    NULL,
    period_end         TIMESTAMP    NULL,
    created_at         TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMP    NOT NULL,

    PRIMARY KEY (id)
);

CREATE INDEX idx_invoices_account_id ON invoices (account_id);
CREATE INDEX idx_invoices_stripe_invoice_id ON invoices (stripe_invoice_id);

CREATE TABLE payments
(
    id                 UUID         NOT NULL,
    account_id         UUID         NOT NULL REFERENCES accounts (id),
    stripe_payment_id  VARCHAR(255) NOT NULL UNIQUE,
    stripe_invoice_id  VARCHAR(255) NULL,
    stripe_customer_id VARCHAR(255) NOT NULL,
    amount             BIGINT       NOT NULL DEFAULT 0,
    currency           VARCHAR(10)  NOT NULL DEFAULT 'usd',
    status             VARCHAR(50)  NOT NULL,
    created_at         TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMP    NOT NULL,

    PRIMARY KEY (id)
);

CREATE INDEX idx_payments_account_id ON payments (account_id);
CREATE INDEX idx_payments_stripe_payment_id ON payments (stripe_payment_id);

CREATE TABLE refunds
(
    id               UUID         NOT NULL,
    account_id       UUID         NOT NULL REFERENCES accounts (id),
    payment_id       UUID         NOT NULL REFERENCES payments (id),
    stripe_refund_id VARCHAR(255) NOT NULL UNIQUE,
    amount           BIGINT       NOT NULL DEFAULT 0,
    currency         VARCHAR(10)  NOT NULL DEFAULT 'usd',
    status           VARCHAR(50)  NOT NULL,
    reason           TEXT         NULL,
    created_at       TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP    NOT NULL,

    PRIMARY KEY (id)
);

CREATE INDEX idx_refunds_account_id ON refunds (account_id);
CREATE INDEX idx_refunds_payment_id ON refunds (payment_id);
