Feature Suggestions

High Value / Low Effort

1. Rate Limiting                                                                                                                                                    
   The middleware stack has no rate limiting. Add per-IP and per-user rate limiting using a token bucket pattern (e.g., golang.org/x/time/rate or Redis-backed sliding
   window). Auth endpoints like POST /auth/login are particularly exposed to brute-force.

2. Access Token Blacklisting / Revocation
   Logout currently only removes the refresh token. If an access token leaks, it remains valid for up to 20 minutes. A Redis-backed JWT blocklist keyed by jti (JWT ID)
   with TTL matching the token expiry would close this gap.

  ---
Medium Value / Medium Effort

3. Hierarchical Roles (Role Inheritance)
   The current RBAC is flat — a user can have many roles, but roles don't inherit from each other. Adding a parent_role_id to the roles table enables hierarchical
   permission inheritance (e.g., admin inherits all editor permissions).

4. Pagination Cursor-Based (Keyset)
   The codebase uses offset-based pagination (page, order packages). For large datasets, keyset/cursor pagination performs significantly better and is stable under
   concurrent writes. This is a refactor of the repo query layer.

  ---
Refactor Suggestions

7. Outbox Relay — Push-Based Instead of Poll-Based
    internal/app/relay/relay.go polls the outbox table at a fixed interval. A PostgreSQL LISTEN/NOTIFY trigger on the outbox table would reduce latency and DB load —
    the relay wakes up only on new rows.

8. Config Validation at Startup
    internal/config loads values via Viper but doesn't validate required fields (e.g., missing Vault address, empty DB host). A config.Validate() step during startup
    would give early, clear errors instead of panics deep in initialization.

  ---
Architectural Observation

The outbox + relay + Kafka pattern is well-implemented, but currently only covers user events. As you add more domains (billing, notifications, etc.), consider a
generic domain event dispatcher at the usecase layer so each new usecase doesn't need to manually wire outbox_service — it could be middleware on the composition
layer.