# Helm Chart - Hyperledger Besu

This is a chart for deploying a [Hyperledger Besu](https://github.com/hyperledger/besu) cluster using the QBFT protocol.

# Design

This deployment is built around a StatefulSet to run multiple Besu validators with a few key features:
* Services are split out into their component parts: RPC, WebSocket, P2P, and a headless service for Besu itself so it can handle connections between nodes on its own.
* All nodes are validators, so each node needs a public and private key. They're both created with `scripts/generate-keys.sh`, but only the private keys are mounted to each pod as a Kubernetes secret. The public keys are used to peer with other nodes.
* Automatic peer discovery - this was an interesting one. It's handled by a sidecar container that reads enode URLs from a ConfigMap and dynamically connects validators. In theory, I think this works as a very basic form of service discovery:
  * Pod IPs change all the time, and this is more challenging to deal with without a load balancer or stable IP fronting the service.
  * Both public and private keys are created ahead of time to make this more deterministic. The public keys are used as part of the enode string required by the `admin_addPeer` API.
  * Each validator runs the sidecar, which formats the validator public keys into the required `enode://PUBLIC_KEY@IP:PORT` format and iterates through each node and attempts to peer (except the validator making the query, of course).
* General configuration of Besu is handled in the main container, while a mounted ConfigMap provides the genesis file (which is also created in `scripts/generate-keys.sh`).

# Improvements
* The peer-connector sidecar could be replaced with static enode configuration if the network topology is absolutely stable. I would probably add static IPs for these nodes in production, or front them with a proper load balancer with TLS or even implement better service discovery (of which I'm sure there are many more elegant solutions!)
* Key management could be improved with external secret management like Vault instead of relying on Kubernetes secrets.