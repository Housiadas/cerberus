Access Token Blacklisting / Revocation:
Logout currently only removes the refresh token. 
If an access token leaks, it remains valid for up to 20 minutes. 
A Redis-backed JWT blocklist keyed by jti (JWT ID) with TTL matching the token expiry would close this gap.

Outbox Relay — Push-Based Instead of Poll-Based;
internal/app/relay/relay.go polls the outbox table at a fixed interval. 
A PostgreSQL LISTEN/NOTIFY trigger on the outbox table 
would reduce latency, and DB load the relay wakes up only on new rows.

Core Directory Refactor Analysis

Inconsistent error-wrap messages — "named_exec_context", "namedexeccontext", "error role create in db". 
Standardize to "<op> <entity>: %w". Effort: S

Large service constructors (user/service.go 7 params, auth/service.go nine fields). 
Only auth uses a Config struct. Optional cleanup. Effort: M

No input validation in Create methods — services trust callers.
Add domain constructors (NewAccount, NewRole). Effort: M

---
Plan: Cron Job Infrastructure + Outbox Cleanup Command

Context

The project has two binaries (cmd/rest for HTTP API, cmd/worker for CLI commands) but no infrastructure for scheduled tasks. We need a third binary cmd/cron to run short-lived scheduled tasks via Kubernetes CronJobs. The first
cron task will clean processed outbox rows older than 7 days from both outbox and email_notification_outbox tables.

Implementation Steps

Step 1: Add sqlc queries for outbox cleanup

File: db/query/outbox.sql — add:
-- name: DeleteProcessedOutbox :execrows
DELETE FROM outbox
WHERE processed_at IS NOT NULL
AND processed_at < @before::timestamp;

File: db/query/notification_outbox.sql — add:
-- name: DeleteProcessedNotificationOutbox :execrows
DELETE FROM email_notification_outbox
WHERE processed_at IS NOT NULL
AND processed_at < @before::timestamp;

Using :execrows to return the count of deleted rows for logging.

Step 2: Add index for cleanup query performance

New migration file: db/migrations/000009_add_outbox_processed_index.up.sql

CREATE INDEX idx_outbox_processed ON outbox (processed_at) WHERE processed_at IS NOT NULL;
CREATE INDEX idx_email_notification_outbox_processed ON email_notification_outbox (processed_at) WHERE processed_at IS NOT NULL;

And the corresponding down.sql to drop both indexes.

The existing partial index idx_outbox_unprocessed covers WHERE processed_at IS NULL — but the cleanup query filters WHERE processed_at IS NOT NULL AND processed_at < X, so it needs a separate index.

Step 3: Run make sqlc to generate Go code

This generates DeleteProcessedOutbox and DeleteProcessedNotificationOutbox methods in db/sqlc/.

Step 4: Create the cron command — internal/cron/

Follow the same pattern as internal/command/ but for cron tasks.

File: internal/cron/cron.go — core types:
- Reuse the same Runner interface pattern from internal/command/command.go
- Config struct with DB config, Log, Tracer
- Cron struct with registry, runners, usage printing
- Runners() returns all available cron runners

File: internal/cron/outbox_cleanup.go — the cleanup runner:
- OutboxCleanupRunner implementing Runner
- Name: outbox-cleanup
- Connects to DB, deletes processed rows older than 7 days from both tables
- Logs the count of deleted rows
- Short-lived: runs and exits

Step 5: Create entry point — cmd/cron/main.go

Follow the same pattern as cmd/worker/main.go:
- Load config, init logger, init telemetry
- Create cron.New(cfg)
- Parse os.Args[1] to select the runner
- Execute and exit

Step 6: Update Dockerfile

File: .docker/app/Dockerfile

Add a build stage for the cron binary alongside the existing rest API build:
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
go build -o cerberus-api cmd/rest/main.go && \
go build -o cerberus-cron cmd/cron/main.go
COPY --from=build_stage /home/cerberus-cron .

Step 7: Update Makefile

Add build targets:
- go/cron/run — run cron locally for testing
- Update k8s/build to also build cron image

Step 8: Add Kubernetes CronJob manifest

File: .kubernetes/app/templates/cronjob-outbox-cleanup.yaml

apiVersion: batch/v1
kind: CronJob
metadata:
name: {{ include "app.fullname" . }}-outbox-cleanup
spec:
schedule: "0 12 * * *"
concurrencyPolicy: Forbid
successfulJobsHistoryLimit: 3
failedJobsHistoryLimit: 3
jobTemplate:
spec:
template:
spec:
containers:
- name: outbox-cleanup
image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
command: ["./cerberus-cron", "outbox-cleanup"]
volumeMounts:
- name: config
mountPath: /app/config.yaml
subPath: config.yaml
restartPolicy: OnFailure
volumes:
- name: config
configMap:
name: {{ include "app.fullname" . }}-config

Files to Create

- cmd/cron/main.go
- internal/cron/cron.go
- internal/cron/outbox_cleanup.go
- db/migrations/000009_add_outbox_processed_index.up.sql
- db/migrations/000009_add_outbox_processed_index.down.sql
- .kubernetes/app/templates/cronjob-outbox-cleanup.yaml

Files to Modify

- db/query/outbox.sql — add DeleteProcessedOutbox query
- db/query/notification_outbox.sql — add DeleteProcessedNotificationOutbox query
- db/sqlc/ — regenerated (via make sqlc)
- .docker/app/Dockerfile — build cron binary too
- Makefile — add go/cron/run target

Verification

1. make sqlc — generates new delete query methods
2. make lint — passes
3. make test — passes
4. go build ./cmd/cron/... — compiles
5. Run locally: go run cmd/cron/main.go outbox-cleanup — verify it connects and runs
