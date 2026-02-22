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
