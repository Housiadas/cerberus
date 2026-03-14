## General

Rate
Limiting:                                                                                                                                             
The middleware stack has no rate limiting. Add per-IP and per-user rate limiting using
a token bucket pattern (e.g., golang.org/x/time/rate or Redis-backed sliding window).
Auth endpoints like POST /auth/login are particularly exposed to brute-force.

Access Token Blacklisting / Revocation:
Logout currently only removes the refresh token. If an access token leaks, it remains valid for up to 20 minutes.
A Redis-backed JWT blocklist keyed by jti (JWT ID) with TTL matching the token expiry would close this gap.

Outbox Relay — Push-Based Instead of Poll-Based;
internal/app/relay/relay.go polls the outbox table at a fixed interval.
A PostgreSQL LISTEN/NOTIFY trigger on the outbox table
would reduce latency, and DB load the relay wakes up only on new rows.

## Billing features

## 1. Webhook Handler

The most critical piece. Stripe calls your endpoint asynchronously for everything — payment succeeded, invoice created, subscription cancelled, card expired.

You **cannot rely on synchronous API responses alone.**

```
POST /webhooks/stripe
```

### Key Events to Handle

- `invoice.paid` → mark invoice paid, extend subscription period
- `invoice.payment_failed` → mark `past_due`, trigger dunning
- `customer.subscription.deleted` → mark subscription cancelled
- `payment_method.detached` → remove from your DB
- `charge.refunded` → create refund record

### Important Requirements

- Always validate the `Stripe-Signature` header
- Process events **idempotently**
- Stripe **retries failed webhooks**, so your handler must tolerate duplicates

---

# 2. Subscription Lifecycle Manager

Handles plan changes, upgrades, downgrades, and cancellations with correct **proration**.

```
POST /subscriptions → create
PATCH /subscriptions/:id → upgrade / downgrade plan
DELETE /subscriptions/:id → cancel (immediate or end of period)
```

---

# 3. Invoice Engine

Responsible for generating invoices at the end of each billing period.

Responsibilities include:

- applying discounts
- calculating tax
- triggering the charge attempt

```
POST /invoices/preview → show what next invoice will look like
POST /invoices/:id/pay → manual retry
POST /invoices/:id/void → void a draft/open invoice
```

---

# 4. Dunning System

Defines what happens when a **payment fails**. This is critical for reducing churn.

### Example Dunning Flow

| Day    | Action                                                      |
|--------|-------------------------------------------------------------|
| Day 0  | Payment fails → email: "Payment failed, please update card" |
| Day 3  | Retry charge                                                |
| Day 7  | Retry charge → email: "Action required"                     |
| Day 14 | Retry charge → email: "Subscription will be cancelled"      |
| Day 21 | Cancel subscription → email: "Subscription cancelled"       |

### Infrastructure Needed

- Background job runner
    - cron
    - Sidekiq
    - Temporal
- Email provider

---

# 5. Tax Calculation
**Do not calculate tax manually.**

### Options
- **Stripe Tax** — simplest if already using Stripe

---

# 6. Invoice PDF Generation

Customers and accountants need **downloadable invoices**.

Each PDF should include:

- Company details
- Customer billing address
- VAT / tax number
- Line items
- Tax breakdown
- Payment reference

### Libraries

- `gotenberg` (Go / Docker-based)

---

# Emails You Must Send

| Trigger                | Email                      |
|------------------------|----------------------------|
| Subscription created   | Welcome + receipt          |
| Invoice paid           | Receipt with PDF           |
| Payment failed         | Alert + update card link   |
| Card expiring soon     | Heads up (30 days before)  |
| Subscription cancelled | Confirmation + offboarding |
| Refund issued          | Confirmation               |
| Trial ending           | Reminder (3 days before)   |

---

# Background Jobs You Need

| Job                             | Frequency            |
|---------------------------------|----------------------|
| Retry failed payments (dunning) | Daily                |
| Check trial expirations         | Daily                |
| Sync Stripe state → your DB     | Daily (safety net)   |
| Detect expiring cards           | Daily                |
| Generate usage invoices         | Monthly / period end |
| Prune stale draft invoices      | Weekly               |

---

# What You Can Offload Entirely

If the above feels like a lot — **it is**.

Depending on your stage, you can offload significant chunks.

| Option             | What It Handles                       | Trade-off                     |
|--------------------|---------------------------------------|-------------------------------|
| Stripe Billing     | Subscriptions, invoices, dunning, tax | Less control, Stripe lock-in  |
| Paddle             | Everything including VAT/tax as MoR   | Higher fees, less flexibility |
| Lago (open source) | Billing engine on top of Stripe       | Self-hosted complexity        |
| Orb                | Usage-based billing                   | Expensive, enterprise-focused |

---

# Practical Advice

Use **Stripe Billing + Stripe Tax**.
This covers **~80% of the billing infrastructure** 
out of the box and lets you focus on building your product.
Build a **custom billing engine later** if Stripe’s pricing model stops fitting your needs.

