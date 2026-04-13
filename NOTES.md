Access Token Blacklisting / Revocation:
Logout currently only removes the refresh token. 
If an access token leaks, it remains valid for up to 20 minutes. 
A Redis-backed JWT blocklist keyed by jti (JWT ID) with TTL matching the token expiry would close this gap.

Outbox Relay — Push-Based Instead of Poll-Based;
internal/app/relay/relay.go polls the outbox table at a fixed interval. 
A PostgreSQL LISTEN/NOTIFY trigger on the outbox table 
would reduce latency, and DB load the relay wakes up only on new rows.

---
Core Directory Refactor Analysis

Inconsistent error-wrap messages — "named_exec_context", "namedexeccontext", "error role create in db". 
Standardize to "<op> <entity>: %w". Effort: S

Large service constructors (user/service.go 7 params, auth/service.go 9 fields). 
Only auth uses a Config struct. Optional cleanup. Effort: M

toXDB/toXDomain mapper boilerplate in every repo — candidate for go:generate. Effort: L

No input validation in Create methods — services trust callers.
Add domain constructors (NewAccount, NewRole). Effort: M

---
sqlc + squirrel + pgxpool Refactor — Cerberus Repo Layer

Context

The cerberus repo layer has 17 _repo packages following a uniform pattern: //go:embed static SQL + pkg/pgsql helpers (NamedExecContext, NamedQueryStruct, NamedQuerySlice, SelectSlice) + hand-written
row structs with db:"..." tags + hand-written toXxxDB / toXxxDomain mappers. 62 static SQL files, 61 pgsql helper call sites.

Squirrel is already integrated (pkg/pgsql/builder.go, used by every dynamic Query method via Builder.Select(...).ToSql()) so sqlc's historical "can't do dynamic WHERE" objection is moot.

This refactor does two things together, because they compose:

1. Adopt sqlc for static CRUD — eliminate embed blocks, hand-written row structs, db:"..." tags, and runtime-only SQL validation. Schema drift becomes a make generate error.
2. Migrate from sqlx.DB to pgxpool.Pool — drop the jmoiron/sqlx abstraction tax. jackc/pgx/v5 is already a transitive dep (used as the driver under sqlx today); promote it to the primary DB type.
   sqlc's pgx/v5 mode emits code directly against pgxpool / pgx.Tx, with native pgtype support (pgtype.UUID, pgtype.Timestamptz, arrays, JSONB).

These two are done in one sweep because sqlc's codegen target (database/sql vs pgx/v5) would otherwise have to be chosen twice, and the generated row types differ between the two modes — doing sqlc
first against database/sql and then re-generating against pgx/v5 would churn every repo twice.

Non-goals (hard constraints)

These are out of scope. Every design choice below is bound by them:

1. No changes to Storer interfaces. Every internal/core/<domain>/ports.go interface keeps its exact signature. sqlc config sets emit_interface: false so sqlc does not generate a parallel Querier
   interface. Mocks regenerate against the same Storer contract — zero churn for callers.
2. No changes to error translation contracts. Store methods return the same domain errors (user.ErrUniqueEmail, user.ErrNotFound, etc.) on the same conditions. The sentinel checks inside Store methods
   change (pgx.ErrNoRows replaces sql.ErrNoRows, *pgconn.PgError with code 23505 replaces the lib/pq-style duplicate detection) but the outgoing error from Store to the service layer is byte-identical.
3. No touching the service layer. Zero edits under internal/core/<domain>/service.go or anything outside _repo packages, pkg/pgsql, cmd/ DB initialization, and integration-test bootstrap. If a change
   would require a service-layer edit, the plan is wrong and must be revised.

If any step below appears to violate these, revisit before proceeding.

Target shape (per repo)

Three layers, same Store wrapper as today:

1. db/ subpackage (sqlc-generated, do-not-edit). Configured with sql_package: pgx/v5. Reads migrations/*.up.sql as schema. Emits a *Queries type bound to a DBTX interface satisfied by both
   *pgxpool.Pool and pgx.Tx. Row types use pgtype.* for nullable columns.
2. filter.go + dynamic Query method (squirrel, hand-written). Squirrel builder unchanged. Scan target changes from hand-written userDB to the sqlc-generated db.User. Execution switches from
   pgsql.SelectSlice (sqlx) to pgx.CollectRows(pool.Query(ctx, sql, args...), pgx.RowToStructByName[db.User]).
3. Store (thin wrapper). Owns *pgxpool.Pool + logger. Per-call builds db.New(pgsql.Conn(ctx, s.pool)) to preserve tx-in-context. Translates pgx.ErrNoRows → user.ErrNotFound, *pgconn.PgError{Code:
   "23505"} → user.ErrUniqueEmail. Owns toDomain mappers (value-object construction: name.Parse, mail.Address — real logic, stays).

Key helper pattern

func (s *Store) q(ctx context.Context) *db.Queries {
return db.New(pgsql.Conn(ctx, s.pool))
}

One line, preserves tx-via-context, zero service-layer churn. pgsql.Conn is rewritten (see below) to return a DBTX (satisfied by *pgxpool.Pool or pgx.Tx) instead of sqlx.ExtContext.

What gets deleted per repo

- //go:embed var blocks for static queries.
- Hand-written row structs (userDB, accountDB, etc.) + all db:"..." tags.
- toUserDB param builders — replaced by small toCreateParams/toUpdateParams helpers mapping domain → sqlc-generated *Params structs (still needed because sqlc doesn't know about value objects like
  name.Name).
- Every pgsql.NamedExecContext / NamedQueryStruct / NamedQuerySlice call site for static queries.

What stays

- pkg/pgsql/builder.go (squirrel builder with $N placeholders — pgx uses the same placeholder style, so no change).
- pkg/pgsql/transaction.go — rewritten to wrap pgx.Tx instead of sqlx.Tx. Public API (Conn, Transactor, NewTransactor, RunInTx) keeps the same names and roughly the same signatures (see below) so the
  service layer is unchanged.
- pkg/pgsql/sqldb.go — rewritten. Open now returns *pgxpool.Pool; StatusCheck uses pool.Ping. Error sentinels re-aliased: ErrDBNotFound = pgx.ErrNoRows, ErrDBDuplicatedEntry detected via
  *pgconn.PgError code 23505 in a helper.
- All Storer interfaces. All service-layer code. All mocks (regenerated, unchanged contract).
- Every dynamic Query method using squirrel.

pgx migration — the pieces that must change outside the repos

This is the "touching pkg/pgsql, cmd/, and test bootstrap" surface. It is bounded and enumerable.

pkg/pgsql/sqldb.go — rewrite

Before:
func Open(cfg Config) (*sqlx.DB, error) { /* sqlx.Connect("pgx", dsn) */ }
func StatusCheck(ctx, *sqlx.DB) error
var ErrDBNotFound = sql.ErrNoRows
var ErrDBDuplicatedEntry = errors.New(...)
// + 4 Named* / Select helpers

After:
func Open(cfg Config) (*pgxpool.Pool, error) {
// pgxpool.ParseConfig(dsn) → pool with MaxConns, lifetimes from cfg
}
func StatusCheck(ctx context.Context, pool *pgxpool.Pool) error {
return pool.Ping(ctx)
}
var ErrDBNotFound = pgx.ErrNoRows
var ErrDBDuplicatedEntry = errors.New("duplicated entry")

// Helper kept for error translation in Store wrappers:
func IsUniqueViolation(err error) bool {
var pgErr *pgconn.PgError
return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

The four Named* / SelectSlice helpers are deleted at the end of the rollout (after all 17 repos are migrated), not up front.

pkg/pgsql/transaction.go — rewrite (preserve public API)

The file currently defines ctxKey, txState{tx *sqlx.Tx}, Conn(ctx, *sqlx.DB) sqlx.ExtContext, Transactor{db *sqlx.DB}, and RunInTx. All of this stays shape-for-shape; only the underlying types swap:

// DBTX is the interface sqlc-generated code binds against. Satisfied by
// *pgxpool.Pool and pgx.Tx. Matches sqlc's default pgx/v5 DBTX.
type DBTX interface {
Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
Query(context.Context, string, ...any) (pgx.Rows, error)
QueryRow(context.Context, string, ...any) pgx.Row
}

type txState struct {
tx   pgx.Tx
done atomic.Bool
}

func Conn(ctx context.Context, pool *pgxpool.Pool) DBTX {
if st, ok := ctx.Value(ctxKey{}).(*txState); ok && !st.done.Load() {
return st.tx
}
return pool
}

type Transactor struct {
log  logger
pool *pgxpool.Pool
}

func NewTransactor(log logger, pool *pgxpool.Pool) *Transactor { ... }

func (t *Transactor) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
tx, err := t.pool.BeginTx(ctx, pgx.TxOptions{})
if err != nil { return fmt.Errorf("begin transaction: %w", err) }

     st := &txState{tx: tx}
     defer func() {
         if st.done.Load() { return }
         if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
             t.log.Errorc(ctx, 5, "pgsql.RunInTx", "rollback error", rbErr)
         }
     }()

     if err := fn(context.WithValue(ctx, ctxKey{}, st)); err != nil {
         return err
     }
     st.done.Store(true)
     if err := tx.Commit(ctx); err != nil {
         return fmt.Errorf("commit transaction: %w", err)
     }
     return nil
}

Public surface that the service layer sees (Transactor.RunInTx(ctx, fn)) is identical. Non-goal #3 holds.

DBTX is defined here (not in a generated file) because every Store.q(ctx) helper references it and every sqlc-generated db.New accepts it. sqlc's default pgx/v5 output declares its own DBTX in each
generated package — we can either (a) use the generated one per-repo, or (b) reference the one in pkg/pgsql. Either works; default to (a) — the generated interface is per-repo and avoids a
cross-package dep from generated code. The Conn helper then returns any satisfying the structural interface (Go is structural on interfaces, so the generated per-repo DBTX accepts whatever Conn returns
as long as the methods match). Concretely: Conn returns pgsql.DBTX, each repo's db.New accepts its own db.DBTX, and the two are method-set-compatible. Verify during pilot that sqlc's generated DBTX
matches exactly; if not, add a one-line adapter.

cmd/ — DB initialization

Every entry point in cmd/ that calls pgsql.Open(cfg) and threads *sqlx.DB through the dependency graph changes to *pgxpool.Pool. This is a mechanical rename at the wiring layer. No logic changes.

Integration test bootstrap — internal/sdk/testutil/apitest/

env.go and dependency.go currently spin up a testcontainers Postgres and hand *sqlx.DB to every NewStore. After the refactor:
- env.go opens *pgxpool.Pool against the testcontainer DSN.
- dependency.go passes *pgxpool.Pool to every NewStore(log, pool).
- Transactor construction updates to pool.

Signatures of NewStore change from (log, *sqlx.DB) to (log, *pgxpool.Pool). This is the one signature change — but it is confined to _repo packages and their direct wiring (cmd/ +
apitest/dependency.go). The service layer does not call NewStore directly, so non-goal #3 is preserved.

sqlc configuration

sqlc.yaml at repo root:

version: "2"
sql:
- engine: postgresql
  schema: migrations
  queries: internal/core
  gen:
  go:
  sql_package: pgx/v5
  emit_interface: false
  emit_json_tags: false
  emit_prepared_queries: false
  emit_exact_table_names: false
  emit_empty_slices: true
  emit_pointers_for_null_types: false

Per-repo output routing: if the top-level config doesn't cleanly route generated files into each repo's db/ subdir, fall back to one sqlc.yaml per repo with out: internal/core/<domain>/<repo>/db.
Decide during the pilot.

Makefile target, added near the mockery target (Makefile:314-320), same docker-based pattern:

sqlc:
docker run --rm -v $(PWD):/src -w /src sqlc/sqlc:latest generat

Wired into generate.

Migration order (pilot → rollout)

The pgx migration in pkg/pgsql + cmd/ + apitest/ is a prerequisite that lands first — it has to be atomic across all 17 repos because Open returns one type. The sqlc adoption is then per-repo.

Phase 0 — pgx pool foundation (one PR)

1. Rewrite pkg/pgsql/sqldb.go — Open returns *pgxpool.Pool, error sentinels re-aliased, add IsUniqueViolation helper. Keep all four Named*/SelectSlice helpers temporarily, but reimplement them on top
   of pgx so existing repo code compiles without changes. (They become thin adapters: NamedExecContext builds the SQL with sqlx.Named, then calls pool.Exec. This is temporary ugly code that exists for
   exactly the duration of phases 1–8.)
2. Rewrite pkg/pgsql/transaction.go — Conn, Transactor, RunInTx on pgx.
3. Update cmd/ wiring: *sqlx.DB → *pgxpool.Pool.
4. Update internal/sdk/testutil/apitest/{env.go,dependency.go} wiring.
5. Every _repo.NewStore signature changes from (log, *sqlx.DB) to (log, *pgxpool.Pool) — internals still use the temporary Named* adapters, so repo.go bodies are unchanged aside from the field type.
6. make mockery && make lint && make test — full suite green before moving on.

Gate: phase 0 must be fully green and merged (or at minimum on a stable branch) before phase 1. The repo bodies are unchanged; only the plumbing under them has swapped. This de-risks the pgx migration
independently of sqlc.

Phase 1–8 — per-repo sqlc adoption

Lowest complexity first, same ordering logic as before:

1. refresh_token_repo — pilot. 4 static queries, no dynamic Query, no filter. Proves: sqlc config, q(ctx) helper, error translation pattern, pgx row scanning, Makefile wiring, mockery compatibility.
   Also fixes the pattern inconsistency the explore surfaced (this repo currently stores sqlx.ExtContext directly — standardize on *pgxpool.Pool + Conn(ctx, pool)).
2. reset_token_repo — 3 queries, same shape. Confirms the pilot generalizes.
3. role_permissions_repo, user_roles_repo — 2 queries each, link tables. Cheap wins.
4. audit_repo, outbox_repo, email_notification_outbox_repo, refund_repo — append-mostly tables.
5. permission_repo, role_repo — CRUD + dynamic Query. First real exercise of the squirrel-scans-into-sqlc-row-type combo on pgx (pgx.CollectRows + pgx.RowToStructByName[db.Role]).
6. user_repo, account_repo — core domain. Highest upstream blast radius; do after the pattern is battle-tested.
7. billing_address_repo, invoice_repo, payment_repo, subscription_repo — billing cluster, similar shapes, batchable.
8. user_roles_permissions_repo — last. Reads the user_roles_permissions view (migration 000006). sqlc handles views fine; verify join semantics.

Phase 9 — cleanup

Delete the temporary Named* / SelectSlice adapters from pkg/pgsql/sqldb.go. Grep-verify zero references in internal/. They existed only to let phase 0 ship independently.

Per-repo mechanical steps (template, post-phase-0)

Using refresh_token_repo as the worked example:

1. Rewrite SQL files in sqlc format. Add -- name: CreateRefreshToken :exec headers. Files stay at internal/core/refresh_token/refresh_token_repo/query/*.sql (minimize diff noise, preserve git history).
2. make sqlc — generates internal/core/refresh_token/refresh_token_repo/db/{db.go,models.go,queries.sql.go}.
3. Rewrite repo.go:
- Drop //go:embed vars.
- Add q(ctx) helper returning *db.Queries.
- Replace each pgsql.NamedExecContext(ctx, s.log, pgsql.Conn(ctx, s.pool), sql, toDB(x)) with s.q(ctx).CreateRefreshToken(ctx, toCreateParams(x)).
- Replace each pgsql.NamedQueryStruct(...) with row, err := s.q(ctx).GetRefreshTokenByToken(ctx, token).
- Error translation: errors.Is(err, pgsql.ErrDBNotFound) keeps working (now aliased to pgx.ErrNoRows). Duplicate-detection switches from errors.Is(err, pgsql.ErrDBDuplicatedEntry) to
  pgsql.IsUniqueViolation(err) — the outgoing domain error is unchanged.
4. Rewrite model.go: delete the hand-written row struct, keep toDomain(row db.RefreshToken) (refresh_token.RefreshToken, error), add small toCreateParams/toUpdateParams helpers. Note: sqlc pgx/v5 mode
   uses pgtype.Text, pgtype.Timestamptz, pgtype.UUID for nullable columns — mapper code handles the .Valid flag instead of sql.NullString.
5. make mockery — Storer interface unchanged, mocks regenerate clean.
6. make lint && make test — unit + integration green.
7. Verify — run the repo's integration test file on both branches, outputs identical.

For repos with a dynamic Query method, add step 3a:

3a. Rewrite Query to scan into the sqlc row type via pgx. Squirrel builder is untouched. Execution switches:

query, args, err := sb.ToSql()
if err != nil { return nil, fmt.Errorf("build query: %w", err) }

rows, err := pgsql.Conn(ctx, s.pool).Query(ctx, query, args...)
if err != nil { return nil, fmt.Errorf("query: %w", err) }

dbRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[db.User])
if err != nil { return nil, fmt.Errorf("collect: %w", err) }

return toUsersDomain(dbRows)

pgx.RowToStructByName[T] requires struct field tags (db:"...") — sqlc emits these by default in pgx/v5 mode, so it works out of the box. This is the point at which the hand-written userDB is deleted:
one row type for both static and dynamic paths.

Critical files

New

- sqlc.yaml (root, or per-repo)
- internal/core/<domain>/<repo>/db/{db.go,models.go,queries.sql.go} × 17 (generated, do-not-edit)

Modified (per repo, phases 1–8)

- internal/core/<domain>/<repo>/repo.go — drop embeds, add q(ctx), replace pgsql calls, update error translation
- internal/core/<domain>/<repo>/model.go — delete row struct, keep domain mapper, add param builders
- internal/core/<domain>/<repo>/query/*.sql — add sqlc -- name: headers

Modified (phase 0, one PR)

- pkg/pgsql/sqldb.go — Open → *pgxpool.Pool, error sentinels, IsUniqueViolation, temporary Named* adapters
- pkg/pgsql/transaction.go — Conn, Transactor, RunInTx on pgx.Tx
- pkg/pgsql/builder.go — unchanged (squirrel $N placeholders work with pgx)
- cmd/** — DB type *sqlx.DB → *pgxpool.Pool in wiring
- internal/sdk/testutil/apitest/env.go — pool setup
- internal/sdk/testutil/apitest/dependency.go — pass pool to NewStore
- Every internal/core/<domain>/<repo>/repo.go — change db *sqlx.DB field to pool *pgxpool.Pool (body unchanged in phase 0 because Named* adapters still exist)
- Makefile — add sqlc target, wire into generate
- go.mod — jmoiron/sqlx dependency removed at end of phase 9; jackc/pgx/v5/pgxpool becomes a direct dep

Modified (phase 9)

- pkg/pgsql/sqldb.go — delete the four temporary Named* / SelectSlice adapters

Unchanged (explicit)

- pkg/pgsql/builder.go
- Every Storer interface in internal/core/<domain>/ports.go
- Every service in internal/core/<domain>/service.go
- All mock files (regenerated, no manual edit)

Existing code to reuse

- pkg/pgsql/transaction.go tx-via-context pattern — ctxKey, txState, Conn, Transactor.RunInTx. The shape is preserved; only the inner types swap from sqlx to pgx.
- pkg/pgsql/builder.go squirrel Builder — untouched, already uses $N placeholders which pgx speaks natively.
- internal/core/user/user_repo/filter.go — reference implementation of the squirrel dynamic-query pattern. account_repo, permission_repo, role_repo, audit_repo, user_roles_permissions_repo already
  follow this shape.
- internal/core/user/user_repo/model.go toUserDomain — reference for the domain-mapper layer that survives the refactor. Value-object parsing (name.Parse, name.ParseNull) is real logic.
- Makefile:314-320 mockery docker pattern — mirror it for the sqlc target.

Verification

Phase 0 gates (pgx foundation, one PR)

- make lint && make test fully green with zero repo-body changes. If these pass, pgx migration under the repo layer is correct and sqlc adoption can begin per-repo.

Per-repo gates (phases 1–8, must pass before next repo)

- make sqlc generates without error (proves SQL parses against the schema)
- make mockery regenerates clean (proves Storer unchanged)
- make lint zero new errors
- make test full suite green

Per-repo behavioral verification

- On a scratch branch, run the target repo's integration test file against master and the refactor branch. Outputs must be identical (same assertions, same query counts if instrumented).
- For user_repo and account_repo: additionally run the full auth flow integration suite — largest blast radius.

Final rollout gate (phase 9)

- make generate && make lint && make test fully green
- Grep: zero references to pgsql.NamedExecContext, NamedQueryStruct, NamedQuerySlice, SelectSlice in internal/
- Grep: zero sqlx imports outside vendor (if vendored) — github.com/jmoiron/sqlx removed from go.mod
- LOC delta in final commit message — target ~40% reduction across the repo layer

Correctness spot-checks

- Rename a column in a scratch migration → make sqlc fails with a clear error (compile-time schema validation)
- Break a query intentionally → make sqlc fails (SQL validation)
- Revert, confirm clean generate
- Integration test with an explicit transaction boundary (Transactor.RunInTx) — confirms Conn(ctx, pool) still routes through the tx and rollbacks still trigger on error

Open questions (resolve before phase 0)

1. Single top-level sqlc.yaml or per-repo configs? Default: try top-level; fall back to per-repo if output routing is messy. Decide during the pilot (phase 1).
2. Keep query/*.sql files in place? Default: yes. Preserves git history on the SQL files, minimizes review noise.
3. Use sqlc's generated DBTX interface per repo, or define one centrally in pkg/pgsql? Default: per repo (sqlc's default). Verify during phase 1 that Conn(ctx, pool) returns a value whose method set
   satisfies the generated DBTX. If structural matching doesn't work cleanly, add a one-line adapter or define a central pgsql.DBTX and reference it via sqlc's overrides.
4. emit_interface: true? No. Storer is the contract. A parallel sqlc-generated Querier would be dead weight and confuse mocking.
5. pgxpool config (MaxConns, MaxConnIdleTime, etc.)? Carry over whatever pgsql.Config currently exposes. If fields don't map 1:1, document the differences in the phase 0 PR description.

Plan approved and saved to /home/housi/.claude/plans/velvet-twirling-newt.md. Ready to start on phase 0 whenever you give the go-ahead.
