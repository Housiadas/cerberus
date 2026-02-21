## Kubernetes deployment files

- Create inside `.kuberenetes` the helm charts.
- Respect the same namespace for all the applications.
- Use istio service mesh.
- Add virtual service and ingress for each application.
- Each helm chart should be in a different add files in the app, postgres, tempo, grafana and vault. 
- All applications should communicate with each other through istio service mesh. 
- Use the same namespace for all the applications namespace: curberus 
- Be sure that helm charts work with minikube.
- Add commands in the Makefile to deploy the applications and run it
- Use minikube
