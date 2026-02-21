# Cerberus
A monitoring system built with Go `v1.26`.

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

## Kubernetes Deployment (Minikube)

### Prerequisites
- [minikube](https://minikube.sigs.k8s.io/docs/start/)
- [Helm](https://helm.sh/docs/intro/install/)
- [istioctl](https://istio.io/latest/docs/setup/getting-started/#download)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)

### Setup
Start minikube, install Istio, and create the `cerberus` namespace:
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
Open the minikube tunnel to expose the Istio ingress gateway:
```bash
make k8s/tunnel
```
Then access:
- **API**: `http://cerberus.local`
- **Grafana**: `http://grafana.cerberus.local`

Add these hosts to `/etc/hosts` pointing to `127.0.0.1` when using minikube tunnel.

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

All services run in the `cerberus` namespace with Istio sidecar injection enabled.

### Useful Links
- [go-chi/chi](https://github.com/go-chi/chi)
- [spf13/viper](https://github.com/spf13/viper)
- [swaggo/swag](https://github.com/swaggo/swag)
- [stretchr/testify](https://github.com/stretchr/testify)
- [testcontainers/testcontainers-go](https://github.com/testcontainers/testcontainers-go)
- [vektra/mockery](https://github.com/vektra/mockery)
- [golang-migrate](https://github.com/golang-migrate/migrate)
