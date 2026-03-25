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

Context

The cerberus project uses hexagonal/clean architecture patterns borrowed from Java/C#. While well-executed, this creates unnecessary ceremony in Go: 4 layers of indirection (handler -> usecase ->       
service -> repo), ~1000+ lines of model translation boilerplate, and ~18 files across 4 packages per domain. Go's philosophy is packages as the unit of design — simple, cohesive, domain-centric.

What's Already Good (Keep)

- Constructor-based DI, zero global state
- Interface-based abstractions (Store interfaces)
- Value objects (name, password, money)
- Private fields + public accessors
- Error wrapping, context propagation
- Embedded SQL, OpenAPI codegen
- pkg/ shared libraries, cmd/ entry points

What Changes

- Eliminate hexagonal layers → one domain package with sub-packages for infrastructure
- Kill the usecase layer (9 of 12 are single-service wrappers, worker already bypasses it)
- Eliminate model translation boilerplate (~1000+ lines of toApp*/toBus*/toDB*)
- Flatten internal/ to domain-centric packages

 ---
Target Structure (Two-Tier Split)

internal/
├── user/                        # User domain
│   ├── user.go                  # Types: User, NewUser, UpdateUser, QueryFilter
│   ├── service.go               # Business logic + tx + event dispatch
│   ├── store.go                 # Store interface
│   ├── user_test.go             # Service/domain tests
│   ├── userdb/                  # PostgreSQL implementation
│   │   ├── store.go             # Store implementation
│   │   ├── model.go             # DB model + converters
│   │   ├── filter.go            # SQL filter building
│   │   ├── order.go             # SQL ordering
│   │   └── query/               # Embedded SQL files
│   └── usercache/               # Cache decorator
│       └── cache.go
│
├── role/                        # Same two-tier pattern
│   ├── role.go, service.go, store.go
│   ├── roledb/
│   └── rolecache/
│
├── permission/                  # Same pattern
│   ├── ...
│   ├── permissiondb/
│   └── permissioncache/
│
├── account/                     # Same pattern (no cache)
│   ├── account.go, service.go, store.go
│   └── accountdb/
│
├── audit/
│   ├── audit.go, service.go, store.go
│   └── auditdb/
│
├── outbox/
│   ├── outbox.go, service.go, store.go
│   └── outboxdb/
│
├── subscription/
│   ├── ...
│   └── subscriptiondb/
│
├── invoice/
│   ├── ...
│   └── invoicedb/
│
├── auth/                        # Cross-domain orchestration
│   ├── auth.go                  # Types (tokens, login/register types)
│   ├── service.go               # Orchestrates user, token, email services
│   └── auth_test.go
│
├── billing/                     # Cross-domain orchestration
│   ├── billing.go
│   └── service.go               # Orchestrates account, subscription, invoice, stripe
│
├── userrole/                    # User-role assignments
│   ├── userrole.go, service.go, store.go
│   └── userroledb/
│
├── roleperm/                    # Role-permission assignments
│   ├── roleperm.go, service.go, store.go
│   └── rolepermdb/
│
├── userroleperm/                # Read-only permissions view
│   ├── userroleperm.go, service.go, store.go
│   └── userrolepermdb/
│
├── refreshtoken/
│   ├── ...
│   └── refreshtokendb/
│
├── resettoken/
│   ├── ...
│   └── resettokendb/
│
├── emailnotif/                  # Email notification outbox
│   ├── ...
│   └── emailnotifdb/
│
├── name/                        # Shared value object (from core/domain/name)
├── password/                    # Shared value object
├── money/                       # Shared value object
├── entity/                      # Entity type enum
├── event/                       # Domain event types
├── eventbus/                    # Event dispatcher
│
├── errs/                        # Error types (from utils/errs)
├── appctx/                      # Context helpers (from utils/context)
│
├── web/                         # HTTP layer
│   ├── handler.go               # Handler struct + DI wiring
│   ├── router.go                # Chi router setup
│   ├── user.go                  # User HTTP handlers
│   ├── role.go                  # Role HTTP handlers
│   ├── auth.go                  # Auth HTTP handlers
│   ├── billing.go               # Billing HTTP handlers
│   ├── ...                      # Other handler files
│   ├── openapi/                 # Generated OpenAPI code
│   └── middleware/              # HTTP middleware
│
├── config/                      # Configuration (stays as-is)
│
└── testutil/                    # Test helpers (from utils/*test*)
├── apitest/
├── dbtest/
├── kafkatest/
├── redistest/
└── unitest/

What Gets Eliminated

- internal/core/ directory — domain + service merge into domain packages
- internal/usecase/ directory — handler calls services directly
- internal/app/repo/ directory — repos become *db/ sub-packages
- internal/app/cache/ directory — caches become *cache/ sub-packages
- internal/utils/ directory — relocated to errs/, appctx/, testutil/
- ~1000+ lines of usecase model translation boilerplate

Layer Reduction

BEFORE: Handler → Usecase → Service → Repo   (4 layers, 4 packages per domain)
AFTER:  Handler → Service → Store             (2 layers, 1 package + sub-packages per domain)

 ---
How Each Layer Transforms

Service Absorbs Usecase Responsibilities

The service gains: transaction management + event dispatching (currently in usecase).

Before (usecase/user_usecase/usecase.go):
func (a *UseCase) Create(ctx context.Context, app NewUser) (User, error) {
nc, err := toBusNewUser(app)           // translation boilerplate
var usr user.User
txErr := pgsql.RunInTx(ctx, ..., func(tran pgsql.CommitRollbacker) error {
userCoreTx, _ := a.userCore.NewWithTx(tran)
usr, err = userCoreTx.Create(ctx, nc)
return a.dispatcher.Dispatch(ctx, tran, ...)
})
return toAppUser(usr), nil             // translation boilerplate
}

After (user/service.go):
func (s *Service) Create(ctx context.Context, nu NewUser) (User, error) {
// Business logic directly — no translation layer
usr := User{...}
err := pgsql.RunInTx(ctx, s.log, s.tx, func(tran pgsql.CommitRollbacker) error {
storerTx, _ := s.storer.NewWithTx(tran)
if err := storerTx.Create(ctx, usr); err != nil { return err }
return s.dispatcher.Dispatch(ctx, tran, ...)
})
return usr, err
}

Handler Uses Domain Types Directly

Before (handler → usecase types → domain types):
func (h *Handler) CreateUser(ctx context.Context, req openapi.CreateUserRequestObject) (...) {
nu := user_usecase.NewUser{Name: *req.Body.Name, ...}  // usecase types
usr, err := h.usecase.user.Create(ctx, nu)              // returns usecase.User
return openapi.CreateUser201JSONResponse{Id: usr.ID}    // from usecase types
}

After (handler → domain types directly):
func (h *Handler) CreateUser(ctx context.Context, req openapi.CreateUserRequestObject) (...) {
nu, err := user.ParseNewUser(*req.Body.Name, *req.Body.Email, ...)  // domain types
usr, err := h.userSvc.Create(ctx, nu)                                // returns user.User
return openapi.CreateUser201JSONResponse{Id: usr.ID().String()}      // from domain types
}

Input validation (string → value object parsing) moves to Parse* functions in the domain package.

 ---
Migration Strategy: Phased, Domain-by-Domain

Each phase keeps the project compiling and tests passing.

Phase 0: Foundation (Shared Packages + Web Layer Shell)

Move shared packages first so domain migrations have a stable base.

1. Move value objects: core/domain/name/ → internal/name/, same for password/, money/, entity/, event/
2. Move utils/errs/ → internal/errs/ (update ~35 imports)
3. Move utils/context/ → internal/appctx/ (rename package to appctx, update ~10 imports)
4. Move test helpers: utils/*test*/ → internal/testutil/ (update test imports)
5. Move app/event_dispatcher/ → internal/eventbus/
6. Create internal/web/ shell — move app/handler/openapi/ and app/middleware/ into it
7. Create internal/web/handler.go that initially wraps/delegates to old handler
8. Delete empty internal/utils/
9. Update all imports, make lint && make test

Files affected: ~60-70 import path changes

Phase 1: Pilot — user Domain

Migrate user as proof-of-concept for the two-tier pattern.

1. Create internal/user/user.go — merge types from core/domain/user/user.go
2. Create internal/user/store.go — Store interface from core/domain/user/ports.go (remove pgsql import, use local CommitRollbacker or keep in store.go)
3. Create internal/user/service.go — merge core/service/user_service/ + absorb tx/event logic from usecase/user_usecase/
4. Create internal/user/userdb/ — move app/repo/user_repo/ files (store.go, model.go, filter.go, order.go, query/)
5. Create internal/user/usercache/ — move app/cache/user_cache/
6. Create internal/web/user.go — handler using user.Service directly (no usecase layer)
7. Update internal/web/handler.go wiring for user
8. Remove old: core/domain/user/, core/service/user_service/, usecase/user_usecase/, app/repo/user_repo/, app/cache/user_cache/
9. make mockery && make lint && make test

Key decisions during pilot:
- Finalize the NewWithTx pattern (domain-level tx interface vs. pgsql.CommitRollbacker)
- Finalize the Parse* input validation pattern
- Establish the template all other domains will follow

Phase 2: Simple Domains (Single-Service Pattern)

Apply the user template to each domain that had a single-service usecase:

Batch 2a (core RBAC): role, permission
Batch 2b (assignments): userrole, roleperm, userroleperm
Batch 2c (infrastructure): audit, outbox, emailnotif
Batch 2d (tokens): refreshtoken, resettoken
Batch 2e (billing entities): account, subscription, invoice

Each batch: create domain package + sub-packages, move code, update web handlers, remove old packages, verify.

Phase 3: Complex Domains (Multi-Service Orchestration)

These are the only packages that justify cross-domain composition:

auth — orchestrates user + refreshtoken + resettoken + emailnotif + userroleperm
- Create internal/auth/service.go with dependencies on other domain services
- Move auth-specific logic from usecase/auth_usecase/
- Create internal/web/auth.go handler

billing — orchestrates account + subscription + invoice + stripe
- Create internal/billing/service.go
- Move from usecase/billing_usecase/
- Create internal/web/billing.go handler

Phase 4: Cleanup

1. Remove empty directories: internal/core/, internal/usecase/, internal/app/repo/, internal/app/cache/
2. Remove old internal/app/handler/ (replaced by internal/web/)
3. Remove internal/app/ if empty
4. Update cmd/rest/main.go and cmd/worker/main.go imports
5. Update .mockery.yaml package paths
6. Final make mockery && make lint && make test

 ---
Key Files to Modify

Wiring (touched in every phase)

- internal/web/handler.go — new DI wiring hub (replaces app/handler/handler.go)
- internal/web/router.go — new router (replaces app/handler/router.go)
- cmd/rest/main.go — entry point imports
- cmd/worker/main.go — worker imports
- .mockery.yaml — mock generation config

Per-Domain Template (using user as example)

┌─────────────────────────────────┬──────────────────────────────────────────────┐
│          Old Location           │                 New Location                 │
├─────────────────────────────────┼──────────────────────────────────────────────┤
│ core/domain/user/user.go        │ internal/user/user.go                        │
├─────────────────────────────────┼──────────────────────────────────────────────┤
│ core/domain/user/ports.go       │ internal/user/store.go                       │
├─────────────────────────────────┼──────────────────────────────────────────────┤
│ core/service/user_service/*.go  │ internal/user/service.go                     │
├─────────────────────────────────┼──────────────────────────────────────────────┤
│ usecase/user_usecase/usecase.go │ absorbed into user/service.go + web/user.go  │
├─────────────────────────────────┼──────────────────────────────────────────────┤
│ usecase/user_usecase/model.go   │ deleted (translation boilerplate eliminated) │
├─────────────────────────────────┼──────────────────────────────────────────────┤
│ usecase/user_usecase/filter.go  │ absorbed into user/user.go or web/user.go    │
├─────────────────────────────────┼──────────────────────────────────────────────┤
│ usecase/user_usecase/order.go   │ absorbed into user/user.go or web/user.go    │
├─────────────────────────────────┼──────────────────────────────────────────────┤
│ app/repo/user_repo/repo.go      │ internal/user/userdb/store.go                │
├─────────────────────────────────┼──────────────────────────────────────────────┤
│ app/repo/user_repo/model.go     │ internal/user/userdb/model.go                │
├─────────────────────────────────┼──────────────────────────────────────────────┤
│ app/repo/user_repo/filter.go    │ internal/user/userdb/filter.go               │
├─────────────────────────────────┼──────────────────────────────────────────────┤
│ app/repo/user_repo/order.go     │ internal/user/userdb/order.go                │
├─────────────────────────────────┼──────────────────────────────────────────────┤
│ app/repo/user_repo/query/*.sql  │ internal/user/userdb/query/*.sql             │
├─────────────────────────────────┼──────────────────────────────────────────────┤
│ app/cache/user_cache/           │ internal/user/usercache/cache.go             │
├─────────────────────────────────┼──────────────────────────────────────────────┤
│ app/handler/user.go             │ internal/web/user.go                         │
└─────────────────────────────────┴──────────────────────────────────────────────┘

 ---
Verification (after every phase)

go build ./...    # compilation check
make mockery      # regenerate mocks
make lint         # code style
make test         # full test suite

 ---
Estimated Impact

┌─────────────────────────────────┬───────────────────┬───────────────────────────────┐
│             Metric              │      Before       │             After             │
├─────────────────────────────────┼───────────────────┼───────────────────────────────┤
│ Packages in internal/           │ ~55               │ ~45                           │
├─────────────────────────────────┼───────────────────┼───────────────────────────────┤
│ Layers of indirection           │ 4                 │ 2                             │
├─────────────────────────────────┼───────────────────┼───────────────────────────────┤
│ Translation boilerplate         │ ~1000+ LOC        │ ~0                            │
├─────────────────────────────────┼───────────────────┼───────────────────────────────┤
│ Files per domain                │ ~18 across 4 pkgs │ ~10 across 1 pkg + 2 sub-pkgs │
├─────────────────────────────────┼───────────────────┼───────────────────────────────┤
│ Import statements in handler.go │ ~55               │ ~25                           │
├─────────────────────────────────┼───────────────────┼───────────────────────────────┤
│ Usecase packages                │ 12                │ 0                             │
└─────────────────────────────────┴───────────────────┴───────────────────────────────┘
