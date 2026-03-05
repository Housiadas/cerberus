## Kubernetes deployment files

- Create inside `.kuberenetes` the helm charts.
- Respect the same namespace for all the applications.
- Each helm chart should be in a different add files in the app, postgres, tempo, grafana and vault.
- Use the same namespace for all the applications namespace: cerberus
- Be sure that helm charts work with minikube.
- Add commands in the Makefile to deploy the applications and run it
- Use minikube
- Use NodePort services for app and grafana to expose them via minikube service

## Unit tests task

- Add unit tests in the service package under core
- Add unit tests like the create_test.go under user_service
- Add tests for all the services
- Make sure that tests are running correctly (PASS)
- Make changes if need.
- Mock extra interfaces with mockery if need (.mockery.yaml)

## Soft deletes
- Add deleted_at field to all SQL files
- Update all domain and services under core
- Update all usecases
- Update all repositories
- Update all sql queries under repo package
- Update all queries like fetch 
where delete_at not Null in order to exluded them from fetching
The goal is to create the soft delete feature

## Transactional Outbox Pattern 
- Create the transactional outbox pattern
- Use tables in the database
- Create migrations
- Create outbox domain
- Create events domain that will have
- Make changes for the user domain
- Use Tx transaction 
- Produce messages to kafka in batched and mark rows as proceed
- Add confluent kafka in compose.yaml for local env
- Make changes to kafka package in pkg if need it
- Use mockery for any changes
- Add tests for all the cases
- Run make lint, make test

## Replace swaggo with oapi-codegen
- Add https://github.com/oapi-codegen/oapi-codegen
- Remove swaggo from go.mod
- Run make generate-api-docs
- Run make lint,
- Add oapi-codegen to tools
- Create yml 
```
  package: rest/main
  generate:
  std-http-server: true
  models: true
  strict-server: true
  output: docs/rest.gen.go
```
- Remove everything related to swaggo and replace with oapi-codegen

## Addition of redis
- Add redis to the compose.yaml
- Add go redis client to the go.mod
- Create a wrapper for redis in pkg
- Generate mocks for redis methods
- Add redis to the env
- Add redis to the config and config.dist
- Refactor the cache directory user_cache package
- Use redis for L2 cache and for L1 cache use sturdyc
- Use WithDistributedStorage from sturdyc package
- The flow: sturdyc -> redis -> db
- Use the flow for fetching endpoints not for delete/update endpoints
- Add tests for redis
- Run mockery
- Run make lint
- Run make test

## Roles and permissions CRUD
- Finish the roles and permissions CRUD
- Fix /api/v1/users/{user_id}/role endpoints to add roles to a user
- Fix /api/v1/roles/{role_id}/permissions endpoints to add  permissions to a role
- Create cache layers like the user_cache, bypass operations like Update and Create
- Add tests for roles and permissions
- Add integration tests for roles and permissions
- Respect usecase as the composition layer
- Respect the flow of the application
- Respect the query directory in the repo directory
- Run make mockery
- Run make lint
- Run make test
- Run make generate-api-docs
- Make all the necessary changes to the application in order for the test to run