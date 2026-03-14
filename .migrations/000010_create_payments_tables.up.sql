-- Description: the actual charge attempt. An invoice can have multiple payment records if retries happen.
-- Status values: pending, succeeded, failed, refunded.
-- provider_txn_id is the Stripe charge ID or equivalent for reconciliation.
CREATE TYPE payment_status AS ENUM ('pending', 'succeeded', 'failed', 'refunded');
CREATE TABLE payments
(
    id                UUID           NOT NULL,
    invoice_id        UUID           NOT NULL,
    payment_method_id UUID           NOT NULL,
    amount_cents      INT            NOT NULL CHECK (amount_cents > 0),
    currency          VARCHAR(3)     NOT NULL,
    status            payment_status NOT NULL,
    paid_at           TIMESTAMP      NULL,
    created_at        TIMESTAMP      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP      NOT NULL,

    PRIMARY KEY (id),
    CONSTRAINT fk_payments_invoice FOREIGN KEY (invoice_id) REFERENCES invoices (id),
    CONSTRAINT fk_payments_payment_method FOREIGN KEY (payment_method_id) REFERENCES payment_methods (id)
);

-- Description: stores tokenized card/bank references from your payment provider (Stripe, Adyen, etc.).
-- Never store raw card data here — only the provider token.
-- The is_default flag drives automatic charge attempts.
CREATE TYPE payment_method_type AS ENUM ('card', 'bank_transfer', 'sepa', 'paypal');
CREATE TABLE payment_methods
(
    id         UUID                NOT NULL,
    account_id UUID                NOT NULL,
    type       payment_method_type NOT NULL,
    details    JSONB               NOT NULL DEFAULT '{}',
    is_default BOOLEAN             NOT NULL DEFAULT false,
    created_at TIMESTAMP           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP           NOT NULL,

    PRIMARY KEY (id),
    CONSTRAINT fk_payment_methods_account FOREIGN KEY (account_id) REFERENCES accounts (id)
);

CREATE UNIQUE INDEX idx_payment_methods_default ON payment_methods (account_id) WHERE is_default = true;

-- Description: refunds are issued when a customer asks for a refund.
CREATE TYPE refund_status AS ENUM ('pending', 'succeeded', 'failed');
CREATE TABLE refunds
(
    id           UUID          NOT NULL,
    payment_id   UUID          NOT NULL,
    amount_cents INT           NOT NULL CHECK (amount_cents > 0),
    reason       TEXT          NULL,
    status       refund_status NOT NULL,
    created_at   TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP     NOT NULL,

    PRIMARY KEY (id),
    CONSTRAINT fk_refunds_payment FOREIGN KEY (payment_id) REFERENCES payments (id)
);
