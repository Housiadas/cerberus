# Claude.md
This document describes the project structure and architecture for claude code.

## Stack
- Go
- PostgreSQL
- Docker
- Docker Compose
- Hashicorp Vault
- Kafka
- Redis
- OpenTelemetry
- Grafana Tempo
- Grafana
- Prometheus

## Project Structure

- `.docker` holds docker related files
- `.kubernetes` holds kubernetes related files
- `.migrations` holds database migrations
- `bruno` holds the bruno collections for the API client
- `cmd` holds the application entry point
- `openapi` holds openapi documentation for the openapi-codegen
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
- `usecases`, is the composition layer for the different domain areas

## Rules
- Never mix core/services, for example, never import role_service in the user_service
- The usecase in the composition layer should be the only one that imports different services from the core layer
- Never update mock.go files that are generated from mockery
- Never edit vendor files

# Config file
- config.ymal using the viper library
- Add changes to config.yml and config.ymal.dist for any extra configuration

# Code style
- Respect golangci linters and formatters that are available in the golangci.yaml file
- Use wrapped static errors instead of fmt.Errorf
- Avoid inline error handling

## Testing instructions
- Run `make mockery` to generate mocks for interfaces
- Run `make lint` to run linters and check for code style errors
- Run `make test` to run all the tests
- Fix any test or type errors until the whole suite is green
- Add or update tests for the code you change, even if nobody asked
