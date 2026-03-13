Rate Limiting:                                                                                                                                             
The middleware stack has no rate limiting. Add per-IP and per-user rate limiting using 
a token bucket pattern (e.g., golang.org/x/time/rate or Redis-backed sliding window). 
Auth endpoints like POST /auth/login are particularly exposed to brute-force.

Access Token Blacklisting / Revocation:
Logout currently only removes the refresh token. If an access token leaks, it remains valid for up to 20 minutes. 
A Redis-backed JWT blocklist keyed by jti (JWT ID) with TTL matching the token expiry would close this gap.

Pagination Cursor-Based (Keyset):
The codebase uses offset-based pagination (page, order packages). For large datasets, 
keyset/cursor pagination performs significantly better and is stable under
concurrent writes. This is a refactored of the repo query layer.

Outbox Relay — Push-Based Instead of Poll-Based;
internal/app/relay/relay.go polls the outbox table at a fixed interval. 
A PostgreSQL LISTEN/NOTIFY trigger on the outbox table 
would reduce latency, and DB load the relay wakes up only on new rows.
