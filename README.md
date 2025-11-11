# Racecource Operator

The goal of this challenge project is to deploy a private blockchain for use with a simple Web3 application that is managed by a Kubernetes operator.

# Requirements

You'll need these tools if you don't already have them:
  * colima
  * kind
  * Kubectl
  * helm
  * docker
  * jq

# Deployment

From the root context, run `./scripts/build.sh`. This will bootstrap everything including the Helm charts, operator, keys, secrets, and a simple Racecourse deployment.

Conversely, you can tear down the whole environment with `./scripts/destroy.sh`. Since everything is running in Kubernetes it will just delete the cluster and the Colima VM.

# Design

## The Helm Charts

For this project, I created a few Helm charts to make deployments more straightforward. You'll find their READMEs below:

* [besu](helm/besu/README.md)
* [firefly-signer](helm/firefly-signer/README.md)
* [nginx](helm/nginx/README.md)

## The Operator

The Racecourse Operator was scaffolded by Kubebuilder and manages several basic Kubernetes resources (Deployment, Service, Configmap, and Ingress). There are two examples of the Racecourse CRDs in `operator/config/samples`: minimal and production-grade. The minimal example works locally, but the production-grade example requires some additional tooling and infrasatructre (TLS, image repo) to be fully operational.

## The Application

I had to make some modifications to the Racecourse application in order to get it to function properly, so I forked it into this repo.
* In `server.js`, I added some environment variables in order to bypass the login screen while testing the application. Since we're not using basic auth, it seemed unnecessary to keep punching in the host path.
* In `Racecoure.js`, I was forced to bypass some assertions that were causing the app to crash completely after loading the contract. Apparently this was due to an issue with the way what Web3 (or this very old version of it) handles filters. In their place I added some basic logic to better handle unexpected events (basically null results or crashes). The instructions did say that major dependency upgrades weren't necessary, so I tried to modify this as little as possible.