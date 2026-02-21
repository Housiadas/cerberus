# Claude.md
This document describes the project structure and architecture for claude code.

## Stack
- Go
- PostgreSQL
- Docker
- Docker Compose
- Hashicorp Vault

## Project Structure

- `.docker` holds docker related files
- `.kubernetes` holds kubernetes related files
- `.migrations` holds database migrations
- `cmd` holds the application entry point
- `docs` holds swagger documentation
- `internal` holds the project logic
- `pkg`, holds shared libraries that are not specific to any project
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

# Config file
- config.ymal using the viper library
- Add changes to config.ymal.dist for any extra configuration

# Code style
- Respect golangci linters and formatters that are available in the golangci.yaml file
- Use wrapped static errors instead of fmt.Errorf
- Avoid inline error handling

## Testing instructions
- Run `make lint` to run linters and check for code style errors
- Run `make test` to run all the tests
- Fix any test or type errors until the whole suite is green
- Add or update tests for the code you change, even if nobody asked

### Unit testing
- Run `make mockery` to generate mocks for interfaces that are not mocked yet to create unit tests

## PR instructions
- Create a PR in the project's repository
