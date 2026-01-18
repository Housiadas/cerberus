# ============= #
# VARIABLES
# ============= #

UID			:= $(shell id -u)
GID			:= $(shell id -g)
GO_VERSION	:= 1.25.3

INPUT			?= $(shell bash -c 'read -p "Insert name: " name; echo $$name')
INPUT_TOOL		?= $(shell bash -c 'read -p "Insert tool: " name; echo $$name')
CURRENT_TIME	:= $(shell date --iso-8601=seconds)
GIT_VERSION		:= $(shell git describe --always --dirty --tags --long)
LINKER_FLAGS	:= "-s -X main.buildTime=${CURRENT_TIME} -X main.version=${GIT_VERSION}"

DOCKER_COMPOSE_LOCAL	:= docker compose -f ./compose.yaml
MIGRATION_DB_DSN 		:= "postgres://housi:secret123@db:5432/housi_db?sslmode=disable"
MIGRATE 				:= $(DOCKER_COMPOSE_LOCAL) run --rm migrate

.PHONY: help
help:
	@echo Usage:
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## ================== #
## Docker
## ================== #

## docker/build: Build the application
.PHONY: docker/build
docker/build:
	docker build --target application \
		-t banking-api:local \
		--build-arg GO_VERSION=$(GO_VERSION) \
 		-f .docker/app/Dockerfile .

## docker/up: Start all the containers for the application
.PHONY: docker/up
docker/up:
	make docker/down
	$(DOCKER_COMPOSE_LOCAL) up -d

## docker/stop: stop all containers
.PHONY: docker/stop
docker/stop:
	$(DOCKER_COMPOSE_LOCAL) stop

## docker/down: stop and remove all containers
.PHONY: docker/down
docker/down:
	$(DOCKER_COMPOSE_LOCAL) down --remove-orphans

## docker/golang-ci: Run golang-ci through docker
.PHONY: docker/down
docker/golang-ci:
	docker run -t --rm -v $(pwd):/app -w /app golangci/golangci-lint:v2.8.0 golangci-lint run

## docker/clean: docker clean all
.PHONY: docker/clean
docker/clean:
	docker system prune -f  && \
    docker image prune -f && \
    docker volume prune -f

## ================== #
## Rest Application
## ================== #

## go/rest/run: Run main.go locally
.PHONY: go/rest/run
go/rest/run:
	go run cmd/rest/main.go

## go/rest/build: build the rest application
.PHONY: go/rest/build
go/rest/build:
	cd cm/rest & \
	go build -ldflags=${LINKER_FLAGS} -o=./rest-api

## ========== #
## Database
## ========== #

## db/migrate/create name=$1: Create new migration files
.PHONY: db/migrate/create
db/migrate/create:
	$(MIGRATE) create -seq -ext=.sql -dir=./database/migrations $(INPUT)

## db/migrate/up: Apply all up database .migrations
.PHONY: db/migrate/up
db/migrate/up:
	$(MIGRATE) -path=./.migrations -database=${MIGRATION_DB_DSN} up

## db/migrate/down: Apply all down database .migrations (DROP Database)
.PHONY: db/migrate/down
db/migrate/down:
	$(MIGRATE) -path=./.migrations -database=${MIGRATION_DB_DSN} down

## ================ #
## Quality Control
## ================ #

## tidy: Tidy
.PHONY: tidy
tidy:
	go mod tidy
	go mod verify

## static: Run static analysis
.PHONY: static
static:
	go tool staticcheck ./...

# security: Check security
.PHONY: security
security:
	go tool govulncheck ./...

## vet: Vet examines Go source code and reports suspicious constructs
.PHONY: vet
vet:
	go vet ./...

## fmt: Formatting with standard library
.PHONY: fmt
fmt:
	go fmt ./...

## fmt/yaml: Formatting yaml files
.PHONY: fmt/yaml
fmt/yaml:
	go tool yamlfmt .

## lint/golangci: Run golangci
.PHONY: lint/golangci
lint/golangci:
	docker run -t --rm \
    -v $(PWD):/app -w /app \
    golangci/golangci-lint:v2.8.0 golangci-lint run

## lint: Run linter
.PHONY: lint
lint: tidy tools/install static security vet lint/golangci

## ================ #
## Tests
## ================ #

## test: Run tests
.PHONY: test
test:
	CGO_ENABLED=1 go test -v -cover -short -race -json -p 4 ./... | go tool tparse --all

## coverage: Inspect coverage
.PHONY: coverage
coverage:
	go test -cover -v -coverpkg=./... ./...
	go test -coverprofile=coverage.out -coverpkg=./... ./...
	grep -Ev "test/|gen/|debug/|dbtest|unitest" coverage.out > filtered.out
	go tool cover -func=filtered.out

## ================== #
## Modules support
## ================== #

## deps/vendor: Vendor dependencies
.PHONY: vendor
deps/vendor:
	go mod tidy
	go mod vendor
	go mod verify

## deps/update: Update dependencies
.PHONY: deps/update
deps/update:
	go get -u -v ./...
	go mod tidy
	go mod vendor

## deps/list: List dependencies
.PHONY: deps/list
deps/list:
	go list -m -u -mod=readonly all

## deps/cache/clean: Clean cache dependencies
.PHONY: deps/cache/clean
deps/cache/clean:
	go clean -modcache

## deps/reset: Reset dependencies
.PHONY: deps/reset
deps/reset:
	git checkout -- go.mod
	go mod tidy
	go mod vendor

## list: List modules
.PHONY: list
list:
	go list -mod=mod all

## ========== #
## Tooling
## ========== #

## tools/get: Get tools
.PHONY: tools/get
tools/get:
	go get --tool $(INPUT_TOOL)

## tools/install: Install tools
.PHONY: tools/install
tools/install:
	go install tool

## tools/list: List all tools
.PHONY: tools/list
tools/list:
	go tool

## tools/update: Update tools
.PHONY: tools/update
tools/update:
	go get -u tool

## ======== #
## Utils
## ======== #

## generate: Go generate command
.PHONY: generate
generate:
	go generate ./...

## swagger: Generate swagger docs
.PHONY: swagger
swagger:
	docker run --rm -v $(PWD):/code ghcr.io/swaggo/swag:v1.16.3 init --g cmd/rest/main.go

## mockery: Generate mocks
.PHONY: mockery
mockery:
	docker run --rm \
	-v "$(shell pwd)":/src \
	-w /src \
	vektra/mockery:3

## metrics: See metrics
.PHONY: metrics
metrics:
	expvarmon -ports="localhost:4010" \
	-vars="build,requests,goroutines,errors,panics,mem:memstats.HeapAlloc,mem:memstats.HeapSys,mem:memstats.Sys"

## grafana: Open grafana
.PHONY: grafana
grafana:
	open http://localhost:3000/

## statsviz: Open statsviz
.PHONY: statsviz
statsviz:
	open http://localhost:4010/debug/statsviz

## kafka/ui: Open kafka ui
.PHONY: kafka/ui
kafka/ui:
	open http://localhost:8080
