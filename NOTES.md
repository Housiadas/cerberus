## Kubernetes deployment files

- Create inside `.kuberenetes` the helm charts.
- Respect the same namespace for all the applications.
- Each helm chart should be in a different add files in the app, postgres, tempo, grafana and vault.
- Use the same namespace for all the applications namespace: cerberus
- Be sure that helm charts work with minikube.
- Add commands in the Makefile to deploy the applications and run it
- Use minikube
- Use NodePort services for app and grafana to expose them via minikube service
