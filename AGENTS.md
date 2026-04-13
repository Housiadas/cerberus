### Plan Mode Default
- Enter plan mode for ANY not-trivial task (3+ steps or architectural decisions)
- Use plan mode for verification steps, not just building
- Write detailed specs upfront to reduce ambiguity

### Self-Improvement Loop
- After ANY correction from the user: update `.claude/lessons.md` with the pattern
- Write rules for yourself that prevent the same mistake
- Ruthlessly iterate on these lessons until the mistake rate drops
- Review lessons at session start for a project

### Verification Before Done
- Never mark a task complete without proving it works
- Diff behavior between main and your changes when relevant
- Ask yourself: "Would a staff engineer approve this?"
- Run tests, check logs, demonstrate correctness

### Demand Elegance (Balanced)
- For non-trivial changes: pause and ask "is there a more elegant way?"
- If a fix feels hacky: "Knowing everything I know now, implement the elegant solution"
- Skip this for simple, obvious fixes. Don't overengineer
- Challenge your own work before presenting it

### Skills usage
- Use skills for any task that requires a capability
- Load skills from `.claude/skills/`
- Invoke skills with natural language
- Each skill is one independent capability

### Subagents usage
- Use subagents liberally to keep the main context window clean
- Load subagents from `.claude/agents/`
- For complex problems, throw more compute at it via subagents
- One task per subagent for focused execution on a given tech stack

## Core Principles
- **Simplicity First**: Make every change as simple as possible. Impact minimal code
- **No Laziness**: Find root causes. No temporary fixes. Senior developer standards

## Development Environment

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
- `migrations` holds database migrations
- `bruno` holds the bruno collections for the API client
- `cmd` holds the application entry point
- `openapi` holds openapi documentation for the openapi-codegen
- `internal` holds the project logic
- `pkg` holds shared code and libraries
- `test` holds integration tests

## Idiomatic Go
Always check the skills and make sure you are using the idiomatic Go patterns.

## Rules
- Never update mock.go files that are generated from mockery
- Never edit vendor files

## Verification 

- Run `make sqlc` to generate sql code
- Run `make mockery` to generate mocks for interfaces
- Run `make lint` to run linters and check for code style errors
- Run `make test` to run all the tests
- Fix any test or type errors until the whole suite is green
- Add or update tests for the code you change, even if nobody asked
