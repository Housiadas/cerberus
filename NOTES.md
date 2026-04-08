Access Token Blacklisting / Revocation:
Logout currently only removes the refresh token. If an access token leaks, it remains valid for up to 20 minutes. 
A Redis-backed JWT blocklist keyed by jti (JWT ID) with TTL matching the token expiry would close this gap.

Outbox Relay — Push-Based Instead of Poll-Based;
internal/app/relay/relay.go polls the outbox table at a fixed interval. 
A PostgreSQL LISTEN/NOTIFY trigger on the outbox table 
would reduce latency, and DB load the relay wakes up only on new rows.

-------
Core Directory Refactor Analysis

HIGH Priority

2. Duplicated filter/cursor/order SQL builders — account_repo/filter.go, audit_repo/filter.go, role_repo/filter.go, etc. each redefine                              
   applyFilter/applyCursor/orderByClause. Extract a generic QueryBuilder into pkg/pgsql. Effort: L
3. Missing clock injection in permission and role services — permission/service.go:69,98 and role/service.go:66,95 call time.Now() directly while                   
   account/user/refresh_token inject a clock. Breaks testability. Effort: S

MEDIUM Priority

5. Duplicated OrderBy* constants & GetDefaultOrderBy() across 8+ packages (account/order.go, user/order.go, role/order.go, ...). Move generic helper to pkg/order.  
   Effort: M
6. Dependency interfaces (generator, clock, transactor, dispatcher) defined inside service.go instead of ports.go — inconsistent across packages. Consolidate.      
   Effort: M
7. Inconsistent error-wrap messages — "named_exec_context", "namedexeccontext", "error role create in db". Standardize to "<op> <entity>: %w". Effort: S
8. Mockery coverage gaps — not all packages have generated mock_test.go. Run a unified make mockery pass. Effort: S

LOW Priority

9. Large service constructors (user/service.go 7 params, auth/service.go 9 fields). 
   Only auth uses a Config struct. Optional cleanup. Effort: M
10. toXDB/toXDomain mapper boilerplate in every repo — candidate for go:generate. Effort: L
11. No input validation in Create methods — services trust callers.
    Add domain constructors (NewAccount, NewRole). Effort: M

Quick wins (do first)

- Inject clock in permission/role services (S)

Want me to implement any of these? I'd suggest starting with the three quick wins in plan mode.  
