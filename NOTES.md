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

Verification

1. docker compose up elasticsearch kafka db — services start
2. make lint — no linting errors
3. make test — tests pass
4. go run cmd/worker/main.go indexer — indexer starts, creates indices
5. Create a user via API → outbox relay produces to Kafka → indexer indexes to ES
6. curl http://localhost:9200/cerberus-users/_search — verify user document
7. curl http://localhost:9200/cerberus-audit/_search — verify audit document
