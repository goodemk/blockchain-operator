# Helm Chart - Hyperledger Firefly

Here's a simple chart for deploying [Hyperledger FireFly Signer](https://github.com/hyperledger/firefly-signer) as a proxy for Hyperledger Besu.

# Design

This deployment used an off-the-shelf deployment of Firefly Signer with a few notable additions:
* The user accounts that are required by Racecourse are provisioned with a `Job` via `test-accounts.yaml` after the Chart components have been installed.
  * Five accounts are created, but more can be added. Just update the `accounts:` block in the `values.yaml` file.
  * New accounts are created every time the code is run, but this could be changed if permanence is preferred. This seemed cleaner to implement for local dev purposes.
* Ingress has been added so that you can reach the deployment at `http://localhost/signer`
* Configuration is pushed to the pod via mounted Configmap.

# Improvements
* Add Kubernetes secrets or use something like Vault for passwords. Not particularly necessary in this scope, but unsecured, plaintext security tokens are not a great idea for prod.