Rate Limiting:                                                                                                                                             
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
Problems with current design:
1. Beginner.Begin() returns CommitRollbacker interface — violates "accept interfaces, return structs"
2. GetExtContext does a runtime type assertion to extract sqlx.ExtContext from CommitRollbacker — fragile, should be compile-time safe
3. CommitRollbacker is an unnecessary abstraction — it hides *sqlx.Tx behind an interface when a concrete wrapper would be better

Proposed fix: Replace CommitRollbacker with a concrete *Tx struct:

---
Refactor: Replace CommitRollbacker interface with concrete *Tx struct

Context

The current transaction abstraction in pkg/pgsql/transaction.go violates the Go idiom "accept interfaces, return structs":
- Beginner.Begin() returns CommitRollbacker (an interface)
- GetExtContext() does a runtime type assertion to extract sqlx.ExtContext — fragile and unnecessary
- CommitRollbacker hides *sqlx.Tx behind an interface when a concrete wrapper is cleaner

Approach

Step 1: Rewrite pkg/pgsql/transaction.go

- Replace CommitRollbacker interface with concrete Tx struct wrapping *sqlx.Tx
- Tx exposes Commit(), Rollback(), and ExtContext() sqlx.ExtContext
- Beginner interface becomes Begin() (*Tx, error)
- DBBeginner.Begin() returns *Tx (concrete)
- RunInTx callback becomes func(*Tx) error
- Remove GetExtContext() function entirely

New Tx struct:
type Tx struct {
tx *sqlx.Tx
}

func (t *Tx) Commit() error              { return t.tx.Commit() }
func (t *Tx) Rollback() error            { return t.tx.Rollback() }
func (t *Tx) ExtContext() sqlx.ExtContext { return t.tx }

Step 2: Remove ErrInvalidTransactorType from pkg/pgsql/error.go

No longer needed since GetExtContext is removed.

Step 3: Update all port/interface files (14 files)

Change NewWithTx(tx pgsql.CommitRollbacker) → NewWithTx(tx *pgsql.Tx) in:
- internal/core/{account,audit,billing_address,email_notification_outbox,invoice,outbox,payment,permission,refund,role,role_permissions,subscription,user,user_roles}/ports.go

Step 4: Update all repo files (14 files)

Change NewWithTx signature and replace pgsql.GetExtContext(tx) → tx.ExtContext() in:
- internal/core/*/[module]_repo/repo.go

Step 5: Update all service files (13 files)

Change NewWithTx(tx pgsql.CommitRollbacker) → NewWithTx(tx *pgsql.Tx) in service methods, and update RunInTx callbacks from func(tran pgsql.CommitRollbacker) → func(tran *pgsql.Tx).

Step 6: Update all cache files (3 files)

- internal/core/{permission,role,user}/[module]_cache/[module]_cache.go
- Update both the inner storer interface and the NewWithTx implementation

Step 7: Update eventbus files (2 files)

- internal/eventbus/eventbus.go — Dispatch(ctx, tran pgsql.CommitRollbacker, ...) → Dispatch(ctx, tran *pgsql.Tx, ...)
- internal/eventbus/nop.go — same change

Step 8: Update dispatcher interfaces in services

Services like user/service.go define local dispatcher interface with pgsql.CommitRollbacker — update to *pgsql.Tx.

Also remove the dead beginner interface in internal/core/user/service.go:48-50.

Step 9: Regenerate mocks

Run make mockery to regenerate all mock files (never edit mock files manually).

Step 10: Verification

- make lint
- make test

Files to modify

Core (2 files):
- pkg/pgsql/transaction.go
- pkg/pgsql/error.go

Ports (14 files):
- internal/core/{account,audit,billing_address,email_notification_outbox,invoice,outbox,payment,permission,refund,role,role_permissions,subscription,user,user_roles}/ports.go

Repos (14 files):
- internal/core/*/[module]_repo/repo.go

Services (13 files):
- internal/core/{account,audit,billing_address,email_notification_outbox,invoice,outbox,payment,permission,refund,role,role_permissions,subscription,user}/service.go

Additional (6 files):
- internal/core/user_roles/service.go
- internal/core/{permission,role,user}/[module]_cache/[module]_cache.go
- internal/eventbus/eventbus.go
- internal/eventbus/nop.go
- internal/core/auth/forgot_password.go