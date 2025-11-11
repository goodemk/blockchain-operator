# Helm Chart - NGINX Ingress Controller

This is just a simple values override for the [NGINX Ingress Controller](https://github.com/kubernetes/ingress-nginx) chart.

# Design

Pretty minimal configuration here:
* I disabled admission webhooks are disabled to simplify local development.
* The service is exposed via NodePort with hostPort enabled so you can hit `localhost` directly.
* This ingress class is set as the default for the cluster, but it's still specified wherever possible.

# Improvements
* For production, you'd want to enable TLS and set up proper certificate management.
* I'd also configure this as a load balancer service in AWS, scale it out a bit for avaiability, and shore up resource allocation.
* Of course, admission webhooks should probably be enabled for validation in a prod environment.
