-- Description: Create table accounts and billing_addresses
CREATE TABLE accounts
(
    id                 UUID         NOT NULL,
    name               TEXT         NOT NULL,
    stripe_customer_id VARCHAR(255) UNIQUE,
    created_at         TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMP    NOT NULL,
    deleted_at         TIMESTAMP    NULL     DEFAULT NULL,

    PRIMARY KEY (id)
);

CREATE INDEX idx_accounts_deleted_at ON accounts (id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounts_stripe_customer_id ON accounts (stripe_customer_id) WHERE stripe_customer_id IS NOT NULL;

CREATE TABLE billing_addresses
(
    id          UUID        NOT NULL,
    account_id  UUID        NOT NULL REFERENCES accounts (id),
    line1       TEXT        NOT NULL,
    line2       TEXT        NULL,
    city        TEXT        NOT NULL,
    state       TEXT        NULL,
    postal_code VARCHAR(20) NOT NULL,
    country     VARCHAR(2)  NOT NULL,
    created_at  TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP   NOT NULL,

    PRIMARY KEY (id)
);

CREATE INDEX idx_billing_addresses_account_id ON billing_addresses (account_id);
