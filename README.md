# Cerberus
A monitoring system built with Go `v1.25`.

## Stack
- Go
- PostgreSQL
- Hashicorp Vault
- Docker
- Docker Compose

## Libraries
- [go-chi/chi](https://github.com/go-chi/chi)
- [spf13/viper](https://github.com/spf13/viper)
- [swaggo/swag](https://github.com/swaggo/swag)
- [stretchr/testify](https://github.com/stretchr/testify)
- [testcontainers/testcontainers-go](https://github.com/testcontainers/testcontainers-go)
- [vektra/mockery](https://github.com/vektra/mockery)
- [golang-migrate](https://github.com/golang-migrate/migrate)

## Architecture Principles
Inspired by Clean Architecture and Hexagonal architecture

- `cmd`, holds the application entry point
- `internal`, holds the business logic
- `pkg`, holds shared code and libraries

The `internal` directory is organized as follows:
- `app`, holds the application logic (adapters), like repositories, handlers, middlewares
- `core`, holds the domain logic

The `usecases` directory is responsible for combaning different domain areas and business rules