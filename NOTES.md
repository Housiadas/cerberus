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

Large service constructors (user/service.go 7 params, auth/service.go nine fields). 
Only auth uses a Config struct. Optional cleanup. Effort: M

toXDB/toXDomain mapper boilerplate in every repo — candidate for go:generate. Effort: L

No input validation in Create methods — services trust callers.
Add domain constructors (NewAccount, NewRole). Effort: M

Detailed Analysis: toXDB/toDomain Mapper Boilerplate — go:generate Candidacy
                                                                                                                                                                                                                                    
---                                                                                                                                                                                                                               
1. Current State

66 hand-written mapper functions across 13 files (~830 lines total):

┌────────────────────────┬───────────────────────────────────┬───────────┬───────┐                                                                                                                                                
│         Layer          │               Files               │ Functions │ Lines │                                                                                                                                                
├────────────────────────┼───────────────────────────────────┼───────────┼───────┤                                                                                                                                                
│ Core (DB ↔ Domain)     │ 12 internal/core/*/mappings.go    │ 49        │ ~630  │
├────────────────────────┼───────────────────────────────────┼───────────┼───────┤
│ Web (Domain → OpenAPI) │ 1 internal/web/handler/mapping.go │ 17        │ ~185  │                                                                                                                                                
└────────────────────────┴───────────────────────────────────┴───────────┴───────┘

Every core package follows the same pattern — a mappings.go file with:
- toCreate*Params — Domain → sqlc create params
- toUpdate*Params — Domain → sqlc update params
- toDomain* — sqlc row → Domain (via New() constructor)
- toDomain*s — slice variant (always identical loop)
- toDBQueryFilter — query filter mapping

  ---                                                                                                                                                                                                                               
2. Pattern Classification

Trivial / Fully Automatable (~24%)

Direct field copies with no transformation. Examples: refresh_token, reset_token, role_permissions, user_roles.

func toDomainRefreshToken(t db.RefreshToken) RefreshToken {                                                                                                                                                                       
return New(t.ID, t.UserID, t.Token, t.ExpiresAt.Time, t.CreatedAt.Time, t.Revoked)                                                                                                                                            
}

Predictable Transformations (~57%)

Field copies + recurring pgtype conversions that follow mechanical rules:

┌──────────────────┬───────────────┬────────────────────────────────────────────────────┐                                                                                                                                         
│     DB Type      │  Domain Type  │                      Pattern                       │
├──────────────────┼───────────────┼────────────────────────────────────────────────────┤
│ pgtype.Timestamp │ time.Time     │ .Time / {Time: t, Valid: true}                     │
├──────────────────┼───────────────┼────────────────────────────────────────────────────┤
│ pgtype.Text      │ string        │ .String / {String: s, Valid: true}                 │                                                                                                                                         
├──────────────────┼───────────────┼────────────────────────────────────────────────────┤
│ pgtype.UUID      │ *uuid.UUID    │ (*uuid.UUID)(&u.Bytes) / {Bytes: *id, Valid: true} │                                                                                                                                         
├──────────────────┼───────────────┼────────────────────────────────────────────────────┤                                                                                                                                         
│ pgtype.Timestamp │ *time.Time    │ if .Valid { &.Time }                               │
├──────────────────┼───────────────┼────────────────────────────────────────────────────┤                                                                                                                                         
│ string           │ name.Name     │ name.MustParse(s) / .String()                      │
├──────────────────┼───────────────┼────────────────────────────────────────────────────┤                                                                                                                                         
│ string           │ entity.Entity │ entity.MustParse(s) / .String()                    │
└──────────────────┴───────────────┴────────────────────────────────────────────────────┘

These are rule-based — a generator that knows the type pairs can emit them.

Complex / Not Automatable (~18%)

Functions with conditional logic, multi-step nullable handling, or custom helpers:
- user.toCreateUserParams — conditional AccountID presence
- user.toNullName — custom name.Null conversion
- audit.toCreateAuditParams — conditional validity flags
- Query filter functions with pointer chains and new() calls

  ---                                                                                                                                                                                                                               
3. The Real Cost

It's not about line count. The real problems are:

a) Maintenance drag: Every schema migration that adds/renames a column requires updating 2-4 mapper functions by hand. Miss one field → silent data loss or stale values.

b) Review noise: PRs that add a new entity include 50-80 lines of mechanical mapping code that reviewers rubber-stamp. It crowds out the actual business logic in diffs.

c) Inconsistency risk: Some mappers use New() constructors (role, permission), others use struct literals. Some handle nullable timestamps as *time.Time, others leave them as pgtype.Timestamp. No enforced consistency.

d) Slice mappers are 100% boilerplate: Every single toDomain*s / toOpenAPI*s function is identical:                                                                                                                               
func toDomainXs(rows []db.X) []X {                        
out := make([]X, len(rows))                                                                                                                                                                                                   
for i, r := range rows { out[i] = toDomainX(r) }                                                                                                                                                                              
return out                                      
}                                                                                                                                                                                                                                 
This pattern repeats 12 times in core + 5 times in the handler. A generic function eliminates all of them.
                                                                                                                                                                                                                                    
---                                                                                                                                                                                                                               
4. go:generate Approach Options

Option A: Generic Slice Mapper (Quick Win, Low Risk)

Add a single generic function to pkg/ or internal/sdk/:

func MapSlice[S, D any](src []S, fn func(S) D) []D {                                                                                                                                                                              
out := make([]D, len(src))                                                                                                                                                                                                    
for i, s := range src { out[i] = fn(s) }
return out                                                                                                                                                                                                                    
}

Impact: Eliminates 17 functions instantly. Zero generation, zero tooling, zero risk.                                                                                                                                              
Effort: XS (30 min)

Option B: Struct Tag-Based Generator (Medium Effort)

Write a custom go:generate tool that reads struct tags or a YAML config and emits mapper functions:

//go:generate cerberus-mapper -src=db.Role -dst=Role -constructor=New

The generator inspects both struct types, matches fields by name, applies conversion rules based on type pairs (pgtype.Timestamp → time.Time = .Time), and emits the function.

Pros: Eliminates ~70% of mapper code. Type-safe. Catches missing fields at generation time.                                                                                                                                       
Cons: Custom tooling to maintain. Complex mappers (user, audit) still need hand-written overrides. The constructor pattern (unexported fields + New()) makes field matching harder — you'd need to match constructor parameter
order, not struct fields.

Effort: L (2-4 days for the generator + migration)

Option C: Interface-Based Approach with go-mapper or goverter

Use an existing mapper generation library like https://github.com/jmattheis/goverter:

// goverter:converter                                                                                                                                                                                                             
type Converter interface {                                
// goverter:map Name | StringToName
ConvertRole(db.Role) Role                                                                                                                                                                                                     
}

Pros: Battle-tested, handles edge cases, active community.                                                                                                                                                                        
Cons: Doesn't work well with the constructor pattern (unexported fields). Would require either exposing fields or adding Set* methods. Adds an external dependency to the build pipeline. Fighting the tool when your domain types
are intentionally opaque.

Effort: L (integration + refactoring domain types)
                                                            
---                                                                                                                                                                                                                               
5. Recommendation

The constructor + unexported fields pattern is the key constraint. Your domain types are intentionally opaque — Role has id, name, etc. as unexported fields with a New() constructor. This is good design for domain integrity,
but it makes automated mapping fundamentally harder because generators can't assign to unexported fields.

Pragmatic approach — two phases:

Phase 1 (XS effort, immediate): Generic MapSlice function
- Kills 17 boilerplate functions across the codebase
- Drop-in replacement, no architecture changes
- Already idiomatic Go 1.21+ with generics

Phase 2 (M effort, optional): Template-based generator for the pgtype ↔ Go conversions
- Focus only on the toCreate*Params and toUpdate*Params functions (Domain → DB direction)
- These are the most mechanical and the most error-prone during migrations
- The toDomain* direction is harder to generate due to constructors, but also less error-prone (compiler catches missing constructor args)
- Use //go:generate with a custom template + text/template or jennifer

Skip code generation for toDomain*: The constructor pattern already gives you compile-time safety — if sqlc adds a field, the New() call won't compile until you update the mapper. This is better than a generator that silently
zero-values new fields.
                                                                                                                                                                                                                                    
---                                                                                                                                                                                                                               
6. Verdict

┌──────────────────────────────┬───────────────────────────────────────────────────────────────────────────────────────────────┐
│            Metric            │                                             Score                                             │                                                                                                  
├──────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────┤
│ Boilerplate severity         │ High — 66 functions, 830 lines                                                                │
├──────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────┤
│ Automatable portion          │ ~40% (slice mappers + trivial params mappers)                                                 │                                                                                                  
├──────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────┤                                                                                                  
│ Risk of manual errors        │ Medium — constructor pattern catches domain-side, but DB-side params can silently miss fields │                                                                                                  
├──────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────┤                                                                                                  
│ ROI of full go:generate      │ Low-Medium — the constructor pattern fights most generators                                   │
├──────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────┤                                                                                                  
│ ROI of generic MapSlice      │ Very High — trivial change, big cleanup                                                       │
├──────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────┤                                                                                                  
│ ROI of params-only generator │ Medium — catches the most error-prone direction                                               │
└──────────────────────────────┴───────────────────────────────────────────────────────────────────────────────────────────────┘

Bottom line: The constructor pattern is doing half the work a generator would do (compile-time safety on the domain side). A generic MapSlice + keeping hand-written mappers for the complex cases is the pragmatic Go-idiomatic  
answer. A full mapper generator would fight your (good) domain design more than it helps.
                     

## Elasticsearch integration
Use your golang skills to implement a search engine with elasticsearch
Add elasticsearch configuration to the config file and config file dist
Add elasticsearch to docker compose
Use the latest version of elasticsearch and elasticsearch-go library
Create a package for elasticsearch under pkg to wrap the elasticsearch library and the configuration
Add otel instrumentation to the elasticsearch library
Create a new index for the user data
Create a new index for the audit data
Create a command under the command package to create the indexer should be able to index the user data and audit data and other in the future
Consume kafka messages and index the data in elasticsearch
Use the elasticsearch-go library to consume the data from elasticsearch
Use bulk indexing to improve the performance
Create a plan before starting the implementation

--- 
──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
Ready to code?

Here is Claude's plan:                                                                                                                                                                                    
╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌
Elasticsearch Integration Plan

Context

Cerberus needs search/indexing capabilities. User data flows through a transactional outbox pattern to Kafka (user-events topic). Audit records are stored directly in the database. We need to index
both into Elasticsearch for search.

Key finding: The Kafka message value is the raw User JSON (via MarshalJSON()). The message key is the aggregate UUID. The event type (user.created, user.updated, user.deleted) is stored in the outbox
table but not sent in the Kafka message. We need to add it as a Kafka header so the indexer can distinguish creates/updates from deletes.

Audit records contain fields (actorID, objEntity, action, message) not present in Kafka messages, so audit indexing requires a separate DB poller.

 ---
Implementation Steps

1. Config: internal/config/elasticsearch.go (NEW)

type Elasticsearch struct {
Hosts    string `validate:"required"`
Username string
Password string `validate:"required"`
Index    string `validate:"required"`
}

2. Config: internal/config/config.go (MODIFY)

Add Elasticsearch Elasticsearch to the Config struct.

3. Config files: config.yaml + config.yaml.dist (MODIFY)

Add username: "elastic" field to the elasticsearch section.

4. Docker Compose: compose.yaml (MODIFY)

- Fix volume: /data → elasticsearch_data:/usr/share/elasticsearch/data
- Fix password: <password> → secret123

5. Elasticsearch package: pkg/elasticsearch/ (NEW — 4 files)

Follow existing pkg/ wrapper patterns (kafka, vault, httpclient).

pkg/elasticsearch/error.go — Sentinel errors (ErrConnectionFailed, ErrBulkIndexFailed, ErrIndexNotFound)

pkg/elasticsearch/config.go — Config struct (Addresses []string, Username, Password string)

pkg/elasticsearch/mapping.go — Index mapping constants:
- UserMapping: id (keyword), name (text+keyword), email (keyword), department (keyword), enabled (boolean), createdAt/updatedAt/deletedAt (date)
- AuditMapping: id (keyword), objId (keyword), objEntity (keyword), objName (text+keyword), actorId (keyword), action (keyword), data (object, enabled:false), message (text), createdAt (date)

pkg/elasticsearch/client.go — Client wrapper:
- New(log, cfg) (*Client, error) — creates ES client with basic auth
- CreateIndex(ctx, name, mapping) error — idempotent index creation, OTel span
- BulkIndex(ctx, indexName, docs []Document) error — uses esutil.BulkIndexer, OTel span
- DeleteDocument(ctx, indexName, docID) error — single doc delete, OTel span
- Ping(ctx) error — health check
- OTel via telemetry.AddSpan() (same pattern as relay.go)

6. Add event type header to relay: internal/sdk/relay/relay.go (MODIFY)

Add a Kafka header event-type to each message so consumers can distinguish event types:

msg := &ckafka.Message{
...
Headers: []ckafka.Header{
{Key: "event-type", Value: []byte(entry.EventType())},
},
}

7. Command struct: internal/command/commands.go (MODIFY)

- Add Indexer = "indexer" constant
- Add Elasticsearch config.Elasticsearch to Config and esConfig to Command

8. Index handler: internal/command/index_handler.go (NEW)

Routes Kafka messages to the correct ES index:
- Reads event-type header from Kafka message
- user.created / user.updated → upsert to {prefix}-users index (message value is user JSON, message key is document ID)
- user.deleted → delete from {prefix}-users index

9. Audit indexer: internal/command/audit_indexer.go (NEW)

DB poller (follows outbox relay pattern):
- Polls audit table for records not yet indexed
- Bulk indexes to {prefix}-audit ES index
- Tracks last indexed timestamp to avoid re-indexing

Note: This requires a small DB migration to add an indexed_at column to the audit table, or we track the last indexed ID/timestamp in application state. Simpler approach: use a cursor-based approach
with created_at > lastIndexedAt and store the watermark in memory (reindexes from scratch on restart — acceptable since ES upserts are idempotent).

10. Indexer command: internal/command/indexer.go (NEW)

Follows the same signal-handling pattern as outbox_relay.go:
- Creates ES client, ensures indices exist
- Starts Kafka consumer for user-events → user indexing
- Starts audit DB poller → audit indexing
- Graceful shutdown on SIGINT/SIGTERM

11. Worker main: cmd/worker/main.go (MODIFY)

- Pass cfg.Elasticsearch to command.Config
- Add command.Indexer case to processCommands()
- Add help text

12. Dependencies

go get github.com/elastic/go-elasticsearch/v8
go mod vendor

 ---
File Summary

┌────────┬───────────────────────────────────┬─────────────────────────────┐
│ Action │               File                │           Purpose           │
├────────┼───────────────────────────────────┼─────────────────────────────┤
│ Create │ internal/config/elasticsearch.go  │ ES config struct            │
├────────┼───────────────────────────────────┼─────────────────────────────┤
│ Modify │ internal/config/config.go         │ Add Elasticsearch to Config │
├────────┼───────────────────────────────────┼─────────────────────────────┤
│ Modify │ config.yaml + config.yaml.dist    │ Add username field          │
├────────┼───────────────────────────────────┼─────────────────────────────┤
│ Modify │ compose.yaml                      │ Fix ES volume + password    │
├────────┼───────────────────────────────────┼─────────────────────────────┤
│ Create │ pkg/elasticsearch/error.go        │ Sentinel errors             │
├────────┼───────────────────────────────────┼─────────────────────────────┤
│ Create │ pkg/elasticsearch/config.go       │ Client config struct        │
├────────┼───────────────────────────────────┼─────────────────────────────┤
│ Create │ pkg/elasticsearch/mapping.go      │ Index mappings              │
├────────┼───────────────────────────────────┼─────────────────────────────┤
│ Create │ pkg/elasticsearch/client.go       │ ES client wrapper with OTel │
├────────┼───────────────────────────────────┼─────────────────────────────┤
│ Modify │ internal/sdk/relay/relay.go       │ Add event-type header       │
├────────┼───────────────────────────────────┼─────────────────────────────┤
│ Modify │ internal/command/commands.go      │ Add constant + ES config    │
├────────┼───────────────────────────────────┼─────────────────────────────┤
│ Create │ internal/command/index_handler.go │ Kafka message → ES routing  │
├────────┼───────────────────────────────────┼─────────────────────────────┤
│ Create │ internal/command/audit_indexer.go │ Audit DB poller → ES        │
├────────┼───────────────────────────────────┼─────────────────────────────┤
│ Create │ internal/command/indexer.go       │ Indexer command             │
├────────┼───────────────────────────────────┼─────────────────────────────┤
│ Modify │ cmd/worker/main.go                │ Wire ES config + command    │
└────────┴───────────────────────────────────┴─────────────────────────────┘

 ---
Verification

1. docker compose up elasticsearch kafka db — services start
2. make lint — no linting errors
3. make test — tests pass
4. go run cmd/worker/main.go indexer — indexer starts, creates indices
5. Create a user via API → outbox relay produces to Kafka → indexer indexes to ES
6. curl http://localhost:9200/cerberus-users/_search — verify user document
7. curl http://localhost:9200/cerberus-audit/_search — verify audit document
