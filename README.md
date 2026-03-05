# Cerberus
A monitoring system built with Go `v1.26`.

### OpenAPI Specification
The OpenAPI specification is located under [openapi](./openapi/openapi.yaml)

## Project Structure

- `.docker` holds docker related files
- `.kubernetes` holds kubernetes related files
- `.migrations` holds database migrations
- `cmd` holds the application entry point
- `docs` holds swagger documentation
- `internal` holds the project logic
- `pkg` holds shared code and libraries
- `test` holds integration tests

## Architectural Principles
Inspired by Clean Architecture and Hexagonal architecture

- `cmd`, holds the application entry points
- `internal`, holds the project logic
- `pkg`, holds shared libraries that are not specific to any project

The `internal` directory is organized as follows:
- `app`, holds the application logic (adapters), like repositories, handlers, middlewares, commands
- `core`, holds the domain logic, separated to domain (models) and services

The `usecases` directory is responsible for combaning different domain areas and business rules

## Development

Spin up containers
```bash
make docker/up
```

Run REST API
```bash
make go/rest/run
```

Run migrations
```bash
make db/migrate/up
```

Destroy containers
```bash
make docker/down
```

## Kubernetes Deployment (Minikube)

### Prerequisites
- [minikube](https://minikube.sigs.k8s.io/docs/start/)
- [Helm](https://helm.sh/docs/intro/install/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)

### Setup
Start minikube and create the `cerberus` namespace:
```bash
make k8s/setup
```

### Build Images
Build the application and migration Docker images inside minikube's Docker daemon:
```bash
make k8s/build
```

### Deploy
Deploy all services (PostgreSQL, Vault, Tempo, Grafana, App) via Helm:
```bash
make k8s/deploy
```

### Check Status
```bash
make k8s/status
```

### Access Services
Use minikube service to access NodePort services:
```bash
minikube service cerberus-app -n cerberus
minikube service cerberus-grafana -n cerberus
```

### Teardown
```bash
make k8s/undeploy
```

### Helm Charts
Each service has its own Helm chart under `.kubernetes/`:

| Chart | Service | Ports |
|-------|---------|-------|
| `postgres/` | PostgreSQL 17.5 | 5432 |
| `vault/` | Hashicorp Vault 1.21 | 8200 |
| `tempo/` | Grafana Tempo | 3200, 4317, 4318 |
| `grafana/` | Grafana 11.6.0 | 3000 |
| `app/` | Cerberus REST API | 4000, 4010 |

All services run in the `cerberus` namespace

## Observability

Cerberus ships with a full OpenTelemetry stack: the application exports traces
and metrics to an **OTel Collector**, which fans them out to **Tempo** (traces)
and **Prometheus** (metrics). **Grafana** provides a unified UI over both.

### Stack overview

| Component | Role | Local port |
|-----------|------|-----------|
| OTel Collector | Receives OTLP from the app, routes to backends | 4317 (gRPC), 4318 (HTTP) |
| Grafana Tempo | Distributed trace storage & query | 3200 |
| Prometheus | Metrics storage & query | 9090 |
| Grafana | Dashboards, trace & metric exploration | 3000 |

### Start the observability stack

The application must point at the collector. In `config.yaml`:

```yaml
collector:
  host: "localhost:4317"
  probability: 1.0        # sample rate — lower in production (e.g. 0.05)
  metricInterval: "30s"
```

### Grafana — http://localhost:3000

Grafana is pre-configured with two datasources (no login required in the
default dev setup):

| Datasource | UID | What it shows |
|------------|-----|---------------|
| Prometheus | `prometheus` | HTTP request rates, durations, cache hit/miss, active requests |
| Tempo | `tempo` | Distributed traces, per-request spans |

**Explore traces**

1. Open **Explore** (compass icon in the left sidebar).
2. Select the **Tempo** datasource.
3. Use **Search** to filter by service name (`cerberus`), HTTP method, status
   code, or trace duration.
4. Click any trace to open the flame graph and see every span — HTTP server,
   database queries, cache lookups.

**Explore metrics**

1. Open **Explore** and select the **Prometheus** datasource.
2. Useful metric names to start with:

| Metric | Description |
|--------|-------------|
| `cerberus_http_server_request_total` | Total HTTP requests (by method, path, status) |
| `cerberus_http_server_request_duration_seconds` | Request latency histogram |
| `cerberus_http_server_active_requests` | In-flight requests gauge |
| `cerberus_cache_hit_total` | In-memory (L1) cache hits |
| `cerberus_cache_miss_total` | In-memory (L1) cache misses |
| `cerberus_cache_distributed_hit_total` | Redis (L2) cache hits |
| `cerberus_cache_distributed_miss_total` | Redis (L2) cache misses |
| `cerberus_cache_size` | Current number of entries in the in-memory cache |
| `cerberus_http_client_request_total` | Outgoing HTTP requests to downstream services |

Example PromQL — HTTP error rate over the last 5 minutes:

```promql
sum(rate(cerberus_http_server_request_total{status=~"5.."}[5m]))
/
sum(rate(cerberus_http_server_request_total[5m]))
```

Example PromQL — p99 request latency:

```promql
histogram_quantile(0.99,
  sum by (le) (rate(cerberus_http_server_request_duration_seconds_bucket[5m]))
)
```

**Correlating traces and metrics**

Grafana links traces to metrics automatically when both datasources are
configured. In a Prometheus panel, click a data point and choose
**View in Tempo** to jump directly to traces from that time window.

### Prometheus — http://localhost:9090

Use the Prometheus UI to run ad-hoc PromQL queries or check scrape targets:

- **Status → Targets** — confirms the OTel Collector scrape is `UP`.
- **Graph** — run any PromQL expression directly.

### Tempo — http://localhost:3200

Tempo exposes an HTTP API for direct trace lookup when needed:

```bash
# Fetch a trace by ID
curl http://localhost:3200/api/traces/<trace-id>
```
