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
                                                                           