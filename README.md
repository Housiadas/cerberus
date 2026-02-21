# Cerberus
A monitoring system built with Go `v1.26`.

## Stack
- Go
- PostgreSQL
- Hashicorp Vault
- Docker
- Docker Compose

## Project Structure

- `.docker` holds docker related files
- `.kubernetes` holds kubernetes related files
- `.migrations` holds database migrations
- `cmd` holds the application entry point
- `docs` holds swagger documentation
- `internal` holds the project logic
- `pkg` holds shared code and libraries
- `test` holds integration tests

## Architecture Principles
Inspired by Clean Architecture and Hexagonal architecture

- `cmd`, holds the application entry points
- `internal`, holds the project logic
- `pkg`, holds shared libraries that are not specific to any project

The `internal` directory is organized as follows:
- `app`, holds the application logic (adapters), like repositories, handlers, middlewares, commands
- `core`, holds the domain logic, separated to domain (models) and services

The `usecases` directory is responsible for combaning different domain areas and business rules

### Useful Links
- [go-chi/chi](https://github.com/go-chi/chi)
- [spf13/viper](https://github.com/spf13/viper)
- [swaggo/swag](https://github.com/swaggo/swag)
- [stretchr/testify](https://github.com/stretchr/testify)
- [testcontainers/testcontainers-go](https://github.com/testcontainers/testcontainers-go)
- [vektra/mockery](https://github.com/vektra/mockery)
- [golang-migrate](https://github.com/golang-migrate/migrate)
