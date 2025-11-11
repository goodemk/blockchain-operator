# Racecourse Operator

A Kubernetes operator that manages the Racecourse application and its dependencies.

# Design

This operator was built with Kubebuilder and manages the lifecycle of Racecourse deployments:
* The operator watches for `Racecourse` custom resources and reconciles the necessary Kubernetes objects:
  * A Deployment manages the Racecourse app itself.
  * A Service to expose app within the cluster.
  * A ConfigMap handles connection details to the wallet service.
  * Optionally, it creates an Ingress resource if ingress is enabled.
* The `Racecourse` CRD allows you to specify the wallet service connection, contract address, replica count, and ingress settings.
* Owner references are set on all managed resources so they're cleaned up when the Racecourse resource is deleted.

# Improvements
* Add better status conditions to track deployment health and readiness.
* Hook Prometheus up to some of the endpoints to get a handle on alerting and metrics. This could also be used as a first pass at HPA scaling on CPU, perhaps replaced later with something like Keda.
* Implement proper finalizers to ensure we clean everything up.
