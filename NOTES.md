Access Token Blacklisting / Revocation:
Logout currently only removes the refresh token. If an access token leaks, it remains valid for up to 20 minutes. 
A Redis-backed JWT blocklist keyed by jti (JWT ID) with TTL matching the token expiry would close this gap.

Outbox Relay — Push-Based Instead of Poll-Based;
internal/app/relay/relay.go polls the outbox table at a fixed interval. 
A PostgreSQL LISTEN/NOTIFY trigger on the outbox table 
would reduce latency, and DB load the relay wakes up only on new rows.
