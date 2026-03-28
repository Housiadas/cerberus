Access Token Blacklisting / Revocation:
Logout currently only removes the refresh token. If an access token leaks, it remains valid for up to 20 minutes. 
A Redis-backed JWT blocklist keyed by jti (JWT ID) with TTL matching the token expiry would close this gap.

Outbox Relay — Push-Based Instead of Poll-Based;
internal/app/relay/relay.go polls the outbox table at a fixed interval. 
A PostgreSQL LISTEN/NOTIFY trigger on the outbox table 
would reduce latency, and DB load the relay wakes up only on new rows.

## Idiomatic Go Restructuring Plan: Package-Centric Architecture

Plan: Eliminate Usecase Layer by Merging into Services

The cerberus project has a usecase layer (internal/usecase/) with 12 packages that sits 
between HTTP handlers and domain services. 9 of 12 usecases are simple service wrappers that add transaction management, 
event dispatching, and model translation boilerplate. 
This layer adds ~1000+ lines of translation code and an unnecessary layer of indirection. 
The goal is to absorb usecase responsibilities into services and have handlers call services directly.

Current Flow

Handler → Usecase (tx + event + model translation) → Service → Store

Target Flow

Handler (input parsing and response mapping) → Service (tx + event + business logic) → Store

Key Decisions

Where do API model types go?

Currently: user_usecase.User (public string fields) is aliased by openapi.User via x-go-type.

Approach: Remove x-go-type from openapi.yaml → let oapi-codegen generate independent structs. 
Handlers explicitly map domain types ↔ generated openapi types. This eliminates ALL model.go files.

What services gain

- pgsql.Beginner (tx) field
- *eventbus.EventDispatcher field
- CUD methods wrap in pgsql.RunInTx + dispatcher.Dispatch
- Query/read methods stay as-is (no tx needed)

What handlers gain

- String→domain parsing (uuid.Parse, name.Parse, etc.) — currently in usecase
- Response mapping: domain accessors → openapi response structs
- Cursor/filter/order parsing — currently in usecase

 ---
Execution Plan

Phase 1: Pilot — user domain

1a. Enhance user service (internal/core/user/service.go)
- Add tx pgsql.Beginner and dispatcher *eventbus.EventDispatcher fields to Service struct
- Update NewService() to accept tx + dispatcher
- Wrap Create() in pgsql.RunInTx + event dispatch (absorb from user_usecase.Create)
- Wrap Update() in RunInTx + event dispatch
- Wrap Delete() in RunInTx + event dispatch
- Read methods (Query, QueryByID, QueryByEmail, Authenticate) stay unchanged
- Add newUserEvent() helper (move from usecase)

1b. Update OpenAPI spec (openapi/openapi.yaml)
- Remove x-go-type / x-go-type-import for: User, NewUser, UpdateUser, UpdateMe
- Let oapi-codegen generate plain structs with matching field names/types

1c. Regenerate OpenAPI code
- Run go generate on handler package to regenerate openapi.gen.go

1d. Update handler (internal/web/handler/user.go)
- Remove user_usecase import
- CreateUser: parse request body fields → user.NewUser (name.Parse, mail.Parse, password.ParseConfirm, etc.), call h.userSvc.Create(), build openapi response from domain accessors
- UpdateUser: parse fields → user.UpdateUser, call service, build response
- UpdateMe: delegate to Update with restricted fields
- DeleteUser: uuid.Parse → service.Delete
- GetUser/GetMe: uuid → service.QueryByID → build response
- ListUsers: parse query params → filter/orderBy/cursor → service.Query → build paginated response
- Move parseFilter(), getOrderByFields(), userFieldExtractor() from usecase into handler (or a handler helper)

1e. Update handler.go DI (internal/web/handler/handler.go)
- Replace usecase.user with userSvc *user.Service
- Update New() to pass tx + dispatcher to user.NewService()
- Remove user_usecase import

1f. Update middleware (internal/web/middleware/authenticate.go)
- AuthenticateBasic() currently uses user_usecase.AuthenticateUser struct
- Change to call user.Service.Authenticate() directly with parsed email
- Update middleware struct to hold *user.Service instead of *user_usecase.UseCase

1g. Update service tests (internal/core/user/service_test.go)
- Add tests for tx + event dispatch in Create/Update/Delete
- Mock pgsql.Beginner and eventbus.EventDispatcher

1h. Delete internal/usecase/user_usecase/ directory

1i. Verify
make mockery && make lint && make test

Phase 2: Simple CRUD domains (same pattern as user)

Apply identical pattern to each. Each batch compiles independently.

Batch 2a — RBAC:
- role — service gains tx+events, handler maps directly
- permission — same pattern

Batch 2b — Assignments:
- user_roles — simple: service gains tx+events (Add/Remove)
- role_permissions — same pattern

Batch 2c — Read-only / Simple:
- user_roles_permissions — no tx/events needed, just remove model translation. Update middleware to use service directly.
- audit — read-only, just remove model translation
- refresh_token — no events, absorb uuid parsing into handler

Batch 2d — Account (no events, has tx):
- account — service gains tx (no event dispatcher)

Batch 2e — System:
- system — infrastructure health checks, move Readiness/Liveness directly. Types (Status, Info) move to handler or a system package.

Phase 3: Orchestrators

3a. auth usecase → auth service
- Currently orchestrates: user_usecase, refresh_token_usecase, user_roles_permissions_usecase, user.Service, reset_token.Service, email_notification_outbox.Service
- After Phase 1+2: these become direct service dependencies
- Create internal/core/auth/ package (or keep as separate package) with Service struct
- Dependencies: *user.Service, *refresh_token.Service, *reset_token.Service, *email_notification_outbox.Service, *user_roles_permissions.Service
- Move JWT logic, Login, Logout, RefreshAccessToken, ForgotPassword, ResetPassword, Validate, GenerateAccessToken
- Move model types (LoginReq, Token, etc.) to handler or keep in auth package
- Update openapi.yaml x-go-type for auth types
- Update middleware to use auth.Service

3b. billing usecase → billing service
- Currently orchestrates: account.Service, subscription.Service, invoice.Service, stripe.Client
- Create internal/core/billing/ package with Service struct
- Move all billing logic (CreateAccountWithStripe, CreateCheckoutSession, etc.)
- Move model types (CheckoutRequest, etc.) to handler
- Update openapi.yaml x-go-type for billing types

Phase 4: Cleanup

1. Delete internal/usecase/ directory entirely
2. Update handler.go: replace usecase struct with direct service references
3. Update openapi.yaml: ensure all x-go-type references point to correct locations
4. Regenerate OpenAPI code
5. Update .mockery.yaml if package paths changed
6. Final make mockery && make lint && make test

 ---
Critical Files

┌─────────────────────────────────────────────┬──────────────────────────────────────────────────┐
│                    File                     │                      Action                      │
├─────────────────────────────────────────────┼──────────────────────────────────────────────────┤
│ internal/core/user/service.go               │ Add tx + dispatcher, wrap CUD methods            │
├─────────────────────────────────────────────┼──────────────────────────────────────────────────┤
│ internal/web/handler/handler.go             │ Replace usecase struct with service refs         │
├─────────────────────────────────────────────┼──────────────────────────────────────────────────┤
│ internal/web/handler/user.go                │ Call service directly, parse inputs, map outputs │
├─────────────────────────────────────────────┼──────────────────────────────────────────────────┤
│ internal/web/middleware/middleware.go       │ Replace usecase refs with service refs           │
├─────────────────────────────────────────────┼──────────────────────────────────────────────────┤
│ internal/web/middleware/authenticate.go     │ Use user.Service + auth.Service directly         │
├─────────────────────────────────────────────┼──────────────────────────────────────────────────┤
│ internal/web/middleware/permission.go       │ Use user_roles_permissions.Service directly      │
├─────────────────────────────────────────────┼──────────────────────────────────────────────────┤
│ openapi/openapi.yaml                        │ Update/remove x-go-type references               │
├─────────────────────────────────────────────┼──────────────────────────────────────────────────┤
│ internal/web/handler/openapi/openapi.gen.go │ Regenerated                                      │
├─────────────────────────────────────────────┼──────────────────────────────────────────────────┤
│ internal/usecase/*                          │ All deleted                                      │
└─────────────────────────────────────────────┴──────────────────────────────────────────────────┘

Verification

After each phase:
go build ./...
make mockery
make lint
make test

---
Both are connection pools, but from different driver ecosystems:

*sqlx.DB (jmoiron/sqlx)

- Wraps the standard database/sql *sql.DB
- Uses database/sql driver interface (e.g., lib/pq or pgx/stdlib)
- Generic — works with any SQL database (Postgres, MySQL, SQLite)
- Connection pool managed by database/sql internals
- Struct scanning via sqlx.StructScan, named queries via sqlx.NamedExec
- All queries go through the database/sql abstraction layer (extra allocations, interface conversions)

*pgxpool.Pool (jackc/pgx)

- Postgres-only, native protocol implementation
- No database/sql layer — talks the Postgres wire protocol directly
- Access to Postgres-specific features: COPY, LISTEN/NOTIFY, custom types, c
- omposite types, arrays, pgx.Batch for pipelining
- Built-in connection pool with health checks, AfterConnect hooks
- Better performance — fewer allocations, no driver.Value boxing
- Native support for pgtype (proper uuid, inet, jsonb, hstore, etc.)

Key tradeoffs

┌─────────────────┬─────────────────────┬───────────────────────────────┐                                                                                                                                                         
│                 │ sqlx + database/sql │            pgxpool            │
├─────────────────┼─────────────────────┼───────────────────────────────┤                                                                                                                                                         
│ Portability     │ Any SQL DB          │ Postgres only                 │
├─────────────────┼─────────────────────┼───────────────────────────────┤
│ Performance     │ Good                │ Better (no abstraction layer) │
├─────────────────┼─────────────────────┼───────────────────────────────┤                                                                                                                                                         
│ PG features     │ Limited             │ Full (COPY, LISTEN, batching) │
├─────────────────┼─────────────────────┼───────────────────────────────┤                                                                                                                                                         
│ Ecosystem       │ Broad (any driver)  │ pgx-specific                  │
├─────────────────┼─────────────────────┼───────────────────────────────┤                                                                                                                                                         
│ Struct scanning │ Built-in (sqlx)     │ Need pgx + scany or manual    │
└─────────────────┴─────────────────────┴───────────────────────────────┘

Your project uses sqlx with pgx as the underlying driver (via pgx/stdlib), 
which is a common middle-ground: you get sqlx ergonomics with pgx as the driver, but miss out on pgx-native features like batching and COPY.      