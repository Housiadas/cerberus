CREATE TYPE account_type AS ENUM('personal', 'team', 'enterprise');

-- Description: accounts
CREATE TABLE accounts
(
    id         UUID         NOT NULL,
    name       TEXT         NOT NULL,
    type       account_type NOT NULL,
    enabled    BOOLEAN      NOT NULL DEFAULT true,
    created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP    NOT NULL,
    deleted_at TIMESTAMP    NULL DEFAULT NULL,

    PRIMARY KEY (id)
);

CREATE INDEX idx_accounts_deleted_at ON accounts (id) WHERE deleted_at IS NULL;

-- Description: Required for tax calculation and invoice PDF generation.
CREATE TABLE billing_addresses
(
    id          UUID        NOT NULL,
    account_id  UUID        NOT NULL UNIQUE,
    line1       TEXT        NOT NULL,
    line2       TEXT        NULL,
    city        TEXT        NOT NULL,
    state       TEXT        NULL,
    postal_code VARCHAR(20) NOT NULL,
    country     VARCHAR(2)  NOT NULL,
    created_at  TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP   NOT NULL,

    PRIMARY KEY (id),
    CONSTRAINT fk_billing_addresses_account FOREIGN KEY (account_id) REFERENCES accounts(id)
);
