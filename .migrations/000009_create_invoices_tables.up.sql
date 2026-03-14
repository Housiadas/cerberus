-- Description: invoices — generated at each billing cycle (or on demand for usage-based billing).
-- Status values: draft, open, paid, void, uncollectible.
-- Linking to both account_id and subscription_id lets you query invoices directly
CREATE TYPE invoice_status AS ENUM ('draft', 'open', 'paid', 'void', 'uncollectible');
CREATE TABLE invoices
(
    id              UUID           NOT NULL,
    account_id      UUID           NOT NULL,
    subscription_id UUID           NULL,
    status          invoice_status NOT NULL,
    currency        VARCHAR(3)     NOT NULL,
    due_date        TIMESTAMP      NULL,
    issued_at       TIMESTAMP      NULL,
    paid_at         TIMESTAMP      NULL,
    created_at      TIMESTAMP      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP      NOT NULL,

    PRIMARY KEY (id),
    CONSTRAINT fk_invoices_account FOREIGN KEY (account_id) REFERENCES accounts (id),
    CONSTRAINT fk_invoices_subscription FOREIGN KEY (subscription_id) REFERENCES subscriptions (id)
);

-- 3. tax_rates
CREATE TABLE tax_rates
(
    id          UUID          NOT NULL,
    name        VARCHAR(100)  NOT NULL,
    percentage  NUMERIC(5, 2) NOT NULL CHECK (percentage >= 0 AND percentage <= 100),
    country     VARCHAR(2)    NOT NULL,
    description TEXT          NULL,
    is_active   BOOLEAN       NOT NULL DEFAULT true,
    created_at  TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP     NOT NULL,

    PRIMARY KEY (id)
);

-- Description: invoice_items — line items inside an invoice.
-- Allows you to itemize charges (base plan, seats, add-ons, usage overages)
CREATE TABLE invoice_items
(
    id               UUID      NOT NULL,
    invoice_id       UUID      NOT NULL,
    description      TEXT      NOT NULL,
    quantity         INT       NOT NULL CHECK (quantity > 0),
    unit_price_cents INT       NOT NULL CHECK (unit_price_cents >= 0),
    tax_rate_id      UUID      NULL,
    total_cents      INT GENERATED ALWAYS AS (quantity * unit_price_cents) STORED,
    created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    CONSTRAINT fk_invoice_items_invoice FOREIGN KEY (invoice_id) REFERENCES invoices (id),
    CONSTRAINT fk_invoice_items_tax_rate FOREIGN KEY (tax_rate_id) REFERENCES tax_rates (id)
);
