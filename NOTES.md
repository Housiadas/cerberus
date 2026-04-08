Access Token Blacklisting / Revocation:
Logout currently only removes the refresh token. 
If an access token leaks, it remains valid for up to 20 minutes. 
A Redis-backed JWT blocklist keyed by jti (JWT ID) with TTL matching the token expiry would close this gap.

Outbox Relay — Push-Based Instead of Poll-Based;
internal/app/relay/relay.go polls the outbox table at a fixed interval. 
A PostgreSQL LISTEN/NOTIFY trigger on the outbox table 
would reduce latency, and DB load the relay wakes up only on new rows.

-------
Core Directory Refactor Analysis

HIGH Priority

Duplicated filter/cursor/order SQL builders — account_repo/filter.go, audit_repo/filter.go, role_repo/filter.go, etc. each redefine                              
applyFilter/applyCursor/orderByClause. Extract a generic QueryBuilder into pkg/pgsql. Effort: L

MEDIUM Priority

Duplicated OrderBy* constants & GetDefaultOrderBy() across 8+ packages (account/order.go, user/order.go, role/order.go, ...). 
Move generic helper to pkg/order. Effort: M

Inconsistent error-wrap messages — "named_exec_context", "namedexeccontext", "error role create in db". 
Standardize to "<op> <entity>: %w". Effort: S


LOW Priority

Large service constructors (user/service.go 7 params, auth/service.go 9 fields). 
Only auth uses a Config struct. Optional cleanup. Effort: M

toXDB/toXDomain mapper boilerplate in every repo — candidate for go:generate. Effort: L

No input validation in Create methods — services trust callers.
Add domain constructors (NewAccount, NewRole). Effort: M

---

sqlc Refactor Analysis — Cerberus

Current state

- 17 repos under internal/core/*/[*_repo], each follows a consistent pattern: sqlx + pgsql helpers (NamedExecContext, NamedQuerySlice, NamedQueryStruct), embedded .sql files in query/, hand-written    
  userDB/toDomain mappers, custom filter.go that builds dynamic WHERE/cursor pagination via bytes.Buffer.
- Connection helper pgsql.Conn(ctx, db) enables tx-via-context.
- Domain types live in the parent package (user.User), repos return domain models — clean separation already in place.
- repo.go files are small (~170 LOC) and largely boilerplate around mappers + error translation (pgsql.ErrDBDuplicatedEntry → user.ErrUniqueEmail).

Pros of moving to sqlc

1. Kill the mapper boilerplate. userDB/toUserDomain/toUsersDomain exist in every repo. sqlc generates typed structs + scanning, eliminating ~30–40% of each repo file.
2. Compile-time SQL validation against migrations/. Today a typo in user_query.sql only blows up at runtime in integration tests. sqlc parses the schema and fails go generate.
3. Typed parameters. No more map[string]any{"limit": ...} or struct{ ID string \db:"id"` }one-offs. GeneratedQueries.GetUserByID(ctx, uuid.UUID)` is self-documenting.
4. Better IDE support. Generated methods get autocompletion and go-to-definition into the SQL.
5. Aligns with your Go-idiomatic preference (per memory): less reflection (sqlx tags), more plain database/sql + generated code. sqlc is the most idiomatic choice in the ecosystem right now.
6. pgx upgrade path. sqlc can emit pgx/v5 directly — better perf, native types (pgtype.UUID, arrays, JSONB), no lib/pq. Cerberus already uses Postgres exclusively, so the abstraction tax of            
   database/sql is wasted.
7. Migrations become the source of truth. sqlc reads migrations/*.up.sql — schema drift between code and DB becomes impossible.

Cons / friction points specific to Cerberus

1. Dynamic queries don't fit sqlc. user_repo.Query builds WHERE clauses conditionally from QueryFilter (5 optional fields) plus cursor pagination with composite (field, id) > (:v, :id) tuples. sqlc has
   no first-class dynamic WHERE support. Options:
   - Write one query per filter combo → combinatorial explosion (2^5 = 32 for users alone).                                                                                                               
   - Use sqlc's sqlc.narg() + COALESCE(sqlc.narg('name'), name) tricks → ugly, slower (no index usage on NULL branches), and the cursor (field, id) > (...) tuple comparison can't be parameterized       
   cleanly because the column changes with orderBy.Field.                                                                                                                                                   
   - Keep dynamic listing methods hand-written with sqlx/squirrel, use sqlc only for the static CRUD. This is the realistic path but means two query layers per repo.
2. Loss of pgsql.Conn(ctx, db) tx pattern. Your tx-in-context helper would need replacing — sqlc's Queries struct holds a DBTX interface; you'd build q.WithTx(tx) at the call site or wrap it. Doable   
   but touches every service.
3. Custom error translation lives in the repo today (ErrDBDuplicatedEntry → user.ErrUniqueEmail). With sqlc you still wrap generated calls in a thin method to translate pg errors — so repos don't      
   disappear, they shrink.
4. Audit/outbox/RBAC views. 000006_create_user_roles_permissions_view.up.sql — sqlc handles views fine, but user_roles_permissions_repo likely does cross-table aggregation that may need :many queries
   with joins; works but verify.
5. Mockery workflow. Today you mock the Storer interface defined per package. sqlc generates a concrete *Queries. You'd either: (a) keep hand-written Storer interfaces wrapping *Queries (still
   mockable, still your current pattern), or (b) use sqlc's generated interface. (a) is closer to current shape, minimal disruption.
6. One more codegen step in make alongside mockery and oggen. Low cost but real.
7. Migration cost is non-trivial. 17 repos × ~6 queries each ≈ 100 SQL files to convert + verify against integration tests. Not hard, but a multi-day mechanical refactor.

Recommendation

Adopt sqlc, but selectively. Concretely:

1. Add sqlc.yaml with engine: postgresql, sql_package: pgx/v5, schema = migrations/.
2. Per repo, generate static queries only: Create, Update, Delete, QueryByID, QueryByEmail, etc. These are ~70% of methods and pure win.
3. Keep Query(filter, order, cursor) hand-written with squirrel (idiomatic builder, much cleaner than bytes.Buffer + string concat). Replace pgsql.NamedQuerySlice with squirrel → pgx.
4. Keep the Store struct per package as a thin wrapper: it owns *sqlc.Queries + the squirrel-based listing method, and continues to translate pg errors to domain errors. Service layer is unchanged.    
   Mocks via existing Storer interface — zero churn upstream.
5. Migrate pgsql.Conn(ctx, db) to a small db.WithTx(ctx, fn) helper that builds Queries from the tx.
6. Pilot on one bounded repo first — refresh_token or reset_token (no dynamic listing, ~4 queries). Measure LOC delta and developer ergonomics before rolling out to user, account, audit.

Net call: worth it. The current sqlx + embedded SQL + hand mappers pattern is the exact problem sqlc was built for, and your repos are uniform enough that the refactor is mechanical. The dynamic-query
limitation is real but bounded — only the Query method per repo is affected, and squirrel handles it more cleanly than what's there today. Expect ~40% LOC reduction across the repo layer and           
elimination of an entire class of runtime SQL errors.

The one thing that would make me not do it: if you're planning to add a second database backend. You're not — Postgres is committed in the stack — so the abstraction sqlx provides is dead weight.  

---
