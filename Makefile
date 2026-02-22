# ============= #
# VARIABLES
# ============= #

UID			:= $(shell id -u)
GID			:= $(shell id -g)
GO_VERSION	:= 1.26

K8S_NAMESPACE 	:= "cerberus"
K8S_DIR 		:= ".kubernetes"
K8S_APP     	:= "cerberus-app"
K8S_TEMPO     	:= "cerberus-tempo"
K8S_VAULT     	:= "cerberus-vault"
K8S_GRAFANA     := "cerberus-grafana"
K8S_POSTGRES    := "cerberus-postgres"

INPUT			?= $(shell bash -c 'read -p "Insert name: " name; echo $$name')
INPUT_TOOL		?= $(shell bash -c 'read -p "Insert tool: " name; echo $$name')

CURRENT_TIME	:= $(shell date --iso-8601=seconds)
GIT_VERSION		:= $(shell git describe --always --dirty --tags --long)
LINKER_FLAGS	:= "-s -X main.buildTime=${CURRENT_TIME} -X main.version=${GIT_VERSION}"

DOCKER_COMPOSE_LOCAL	:= docker compose -f ./compose.yaml
MIGRATION_DB_DSN		:= "postgres://housi:secret123@db:5432/housi_db?sslmode=disable"
MIGRATE					:= $(DOCKER_COMPOSE_LOCAL) run --rm migrate

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

## lint: Run linter
.PHONY: lint
lint: tidy tools/install security vet golangci

## golangci: Run golangci
.PHONY: golangci
golangci:
	docker run -t --rm \
    -v $(PWD):/app -w /app \
    golangci/golangci-lint:v2.10.1 golangci-lint run

## ================ #
## Tests
## ================ #

## test: Run tests
.PHONY: test
test:
	CGO_ENABLED=1 go test -v -cover -short -race -json -p 4 ./... | go tool tparse --all

## coverage/run: Run tests and generate filtered coverage profile
.PHONY: coverage/run
coverage/run:
	go test -coverprofile=coverage.out -coverpkg=./... ./...
	grep -Evf .coverignore coverage.out > filtered.out

## coverage: Per-function coverage summary
.PHONY: coverage
coverage: coverage/run
	go tool cover -func=filtered.out

## coverage/html: Interactive HTML report in browser
.PHONY: coverage/html
coverage/html: coverage/run
	go tool cover -html=filtered.out -o coverage.html
	xdg-open coverage.html

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

## ================== #
## Kubernetes
## ================== #

## k8s/setup: Setup minikube and create namespace
.PHONY: k8s/setup
k8s/setup:
	minikube start --driver=docker --memory=4096 --cpus=2
	kubectl create namespace $(K8S_NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -

## k8s/build: Build Docker images into minikube's Docker daemon
.PHONY: k8s/build
k8s/build:
	eval $$(minikube docker-env) && \
	docker build --target application \
		-t cerberus-app:local \
		--build-arg GO_VERSION=$(GO_VERSION) \
		-f .docker/app/Dockerfile . && \
	docker build \
		-t cerberus-migrate:local \
		--build-arg UID=$(UID) \
		--build-arg GID=$(GID) \
		-f .docker/migrate/Dockerfile .

## k8s/deploy: Deploy all helm charts to minikube
.PHONY: k8s/deploy
k8s/deploy:
	helm upgrade --install $(K8S_POSTGRES) $(K8S_DIR)/postgres -n $(K8S_NAMESPACE) --create-namespace --wait
	helm upgrade --install $(K8S_VAULT) $(K8S_DIR)/vault -n $(K8S_NAMESPACE) --wait
	helm upgrade --install $(K8S_TEMPO) $(K8S_DIR)/tempo -n $(K8S_NAMESPACE) --wait
	helm upgrade --install $(K8S_GRAFANA) $(K8S_DIR)/grafana -n $(K8S_NAMESPACE) --wait
	helm upgrade --install $(K8S_APP) $(K8S_DIR)/app -n $(K8S_NAMESPACE) --wait

## k8s/undeploy: Uninstall all helm charts
.PHONY: k8s/undeploy
k8s/undeploy:
	helm uninstall $(K8S_APP) -n $(K8S_NAMESPACE)
	helm uninstall $(K8S_GRAFANA) -n $(K8S_NAMESPACE)
	helm uninstall $(K8S_TEMPO) -n $(K8S_NAMESPACE)
	helm uninstall $(K8S_VAULT) -n $(K8S_NAMESPACE)
	helm uninstall $(K8S_POSTGRES)  -n $(K8S_NAMESPACE)

## k8s/status: Show status of all pods and services
.PHONY: k8s/status
k8s/status:
	kubectl get pods,svc -n $(K8S_NAMESPACE)

## k8s/tunnel: Open NodePort services via minikube
.PHONY: k8s/tunnel
k8s/tunnel:
	@echo "Access app:     minikube service $(K8S_APP) -n $(K8S_NAMESPACE)"
	@echo "Access grafana: minikube service $(K8S_GRAFANA) -n $(K8S_NAMESPACE)"

## k8s/pods: Get all pods
k8s/pods:
	minikube kubectl -- get pods -A

## k8s/clean: Cleanup minikube
.PHONY: k8s/clean
k8s/clean:
	minikube delete --all --purge
