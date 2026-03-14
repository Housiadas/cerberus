-- Description: the contract between an account and a plan.
-- Tracks the current billing period and cancellation.
-- Status values: active, trialing, past_due, canceled.
CREATE TYPE subscription_status AS ENUM ('trialing', 'active', 'past_due', 'canceled', 'unpaid');
CREATE TABLE subscriptions
(
    id                   UUID                NOT NULL,
    account_id           UUID                NOT NULL,
    plan_id              UUID                NOT NULL,
    status               subscription_status NOT NULL,
    current_period_start TIMESTAMP           NOT NULL,
    current_period_end   TIMESTAMP           NOT NULL,
    created_at           TIMESTAMP           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP           NOT NULL,

    PRIMARY KEY (id),
    CONSTRAINT fk_subscriptions_account FOREIGN KEY (account_id) REFERENCES accounts (id),
    CONSTRAINT fk_subscriptions_plan FOREIGN KEY (plan_id) REFERENCES plans (id),
    CONSTRAINT chk_subscriptions_period CHECK (current_period_end > current_period_start)
);

-- Description: plans table
CREATE TYPE plan_interval AS ENUM ('monthly', 'yearly');
CREATE TABLE plans
(
    id          UUID                NOT NULL,
    name        VARCHAR(100) UNIQUE NOT NULL,
    description TEXT                NULL,
    interval    plan_interval       NOT NULL,
    price_cents INT                 NOT NULL CHECK (price_cents >= 0),
    currency    VARCHAR(3)          NOT NULL,
    is_active   BOOLEAN             NOT NULL DEFAULT true,
    created_at  TIMESTAMP           NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP           NOT NULL,

    PRIMARY KEY (id)
);

-- Description: coupons table
CREATE TYPE coupon_discount_type AS ENUM ('percent', 'fixed');
CREATE TABLE coupons
(
    id              UUID                 NOT NULL,
    code            VARCHAR(50) UNIQUE   NOT NULL,
    discount_type   coupon_discount_type NOT NULL,
    discount_value  INT                  NOT NULL CHECK (discount_value > 0),
    currency        VARCHAR(3)           NULL,
    max_redemptions INT                  NULL,
    times_redeemed  INT                  NOT NULL DEFAULT 0,
    is_active       BOOLEAN              NOT NULL DEFAULT true,
    expires_at      TIMESTAMP            NULL,
    created_at      TIMESTAMP            NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP            NOT NULL,

    PRIMARY KEY (id)
);

-- 11. subscription_discounts
CREATE TABLE subscription_discounts
(
    id              UUID      NOT NULL,
    subscription_id UUID      NOT NULL,
    coupon_id       UUID      NOT NULL,
    applied_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    CONSTRAINT fk_subscription_discounts_subscription FOREIGN KEY (subscription_id) REFERENCES subscriptions (id),
    CONSTRAINT fk_subscription_discounts_coupon FOREIGN KEY (coupon_id) REFERENCES coupons (id)
);

-- 9. usage_records
CREATE TABLE usage_records
(
    id              UUID      NOT NULL,
    subscription_id UUID      NOT NULL,
    description     TEXT      NOT NULL,
    quantity        INT       NOT NULL CHECK (quantity > 0),
    recorded_at     TIMESTAMP NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    CONSTRAINT fk_usage_records_subscription FOREIGN KEY (subscription_id) REFERENCES subscriptions (id)
);
