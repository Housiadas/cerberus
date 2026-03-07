Feature Suggestions

High Value / Low Effort

1. Rate Limiting                                                                                                                                                    
   The middleware stack has no rate limiting. Add per-IP and per-user rate limiting using a token bucket pattern (e.g., golang.org/x/time/rate or Redis-backed sliding
   window). Auth endpoints like POST /auth/login are particularly exposed to brute-force.

2. Password Reset Flow
   No POST /auth/forgot-password or POST /auth/reset-password exists. This requires:
- A short-lived reset token domain entity (similar to refresh_token)
- An email notification event through the outbox pattern (already in place)

3. Audit Trail for RBAC Changes
   Currently, only user events (UserCreated, UserUpdated, UserDeleted) are published. Role assignment/removal, permission changes, and login events should also be
   audited — they're high-value from a compliance perspective.

4. Access Token Blacklisting / Revocation
   Logout currently only removes the refresh token. If an access token leaks, it remains valid for up to 20 minutes. A Redis-backed JWT blocklist keyed by jti (JWT ID)
   with TTL matching the token expiry would close this gap.

  ---
Medium Value / Medium Effort

5. Hierarchical Roles (Role Inheritance)
   The current RBAC is flat — a user can have many roles, but roles don't inherit from each other. Adding a parent_role_id to the roles table enables hierarchical
   permission inheritance (e.g., admin inherits all editor permissions).

6. Pagination Cursor-Based (Keyset)
   The codebase uses offset-based pagination (page, order packages). For large datasets, keyset/cursor pagination performs significantly better and is stable under
   concurrent writes. This is a refactor of the repo query layer.

7. User Self-Service Endpoints
   There's no GET /me, PUT /me, or PUT /me/password endpoint. Currently, users can only be managed by privileged callers. A self-service profile endpoint with
   different permission scope would be useful.

8. Multi-Tenancy / Namespacing
   The config.App.Namespace field exists but doesn't appear to be used for data isolation. If this is a multi-tenant SaaS, scoping all domain queries by namespace_id
   (or tenant_id) at the repo layer would enable it.

  ---
Refactor Suggestions

9. Access Token Claims — Roles as IDs, Not Names
   JWT claims currently embed role names as strings. If a role is renamed, existing valid tokens carry the old name. Embedding role IDs and resolving them at
   permission check time (with the existing singleflight cache) is more robust.

10. Outbox Relay — Push-Based Instead of Poll-Based
    internal/app/relay/relay.go polls the outbox table at a fixed interval. A PostgreSQL LISTEN/NOTIFY trigger on the outbox table would reduce latency and DB load —
    the relay wakes up only on new rows.

11. Config Validation at Startup
    internal/config loads values via Viper but doesn't validate required fields (e.g., missing Vault address, empty DB host). A config.Validate() step during startup
    would give early, clear errors instead of panics deep in initialization.

12. Structured Error Responses
    The OpenAPI spec likely defines error response shapes, but adding a consistent ErrorCode field (machine-readable, e.g., "user.not_found") alongside the HTTP status
    would make client error handling more reliable than parsing message strings.

  ---
Architectural Observation

The outbox + relay + Kafka pattern is well-implemented, but currently only covers user events. As you add more domains (billing, notifications, etc.), consider a
generic domain event dispatcher at the usecase layer so each new usecase doesn't need to manually wire outbox_service — it could be middleware on the composition
layer.