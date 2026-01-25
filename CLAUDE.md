# Stack
- Go
- PostgreSQL
- Docker
- Docker Compose

# Config file
- config.ymal using the viper library
- Add changes to config.ymal.dist for any extra configuration

# Code style
- Respect golangci linters and formatters that are available in the golangci.yaml file
- Use wrapped static errors instead of fmt.Errorf
- Avoid inline error handling

# Architecture Principles
Inspired by Clean Architecture and Hexagonal architecture

- `cmd`, holds the application entry point
- `internal`, holds the business logic
- `pkg`, holds shared code and libraries

The `internal` directory is organized as follows:
- `app`, holds the application logic (adapters), like repositories, handlers, middlewares
- `core`, holds the domain logic

The `usecases` directory is responsible for combaning different domain areas and business rules
