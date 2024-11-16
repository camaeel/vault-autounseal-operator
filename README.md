# vault-autounseal-operator
Vault operator for managing vault clusters running in Kubernetes. Operator handles:
* automated initialization of new clusters
* automated unsealing of pods in a clusters
  
Operator assumes vault is deployed using official Hashicorp Vault helm chart.

## Planned features:
* upgrading statefulset pods in graceful manner
* rotating vault pods if TLS certificate is updated

## Installation 

### Prerequisites

Vault installed in the cluster. Important parts of the configuration:
* if using ha mode with raft storage `retry_join` block is configured for auto joining the cluster. `.server.standalone` is set to `false` then.

Example configuration (for kind cluster) can be found in `manifests/vault-values.yml`

### Installation using helm

First add helm repository:
```shell
helm repo add vault-autounseal-operator https://camaeel.github.io/vault-autounseal-operator
helm repo update
```

Install:
```shell
helm upgrade --install vault-autounseal-operator vault-autounseal-operator/vault-autounseal-operator 
```

 

## Algorithm:
1. build vault client
2. get pod seal & init status - https://localhost:8200/v1/sys/seal-status
3. if !initialized
   1. check if init secret is not there
   2. sync (create lease or lock)
   3. call sys/initialize
   4. create secret - unseal keys
   5. create secret - root token
4. if sealed
   1. get secret - unseal keys
   2. call sys/unseal

## Development

### Setting local environment

#### Create kind cluster (or use any other cluster). 

It is easiest to use `Makefile` run `make kind` or alternatively
```shell
kind create cluster \
  --wait 120s \
  --config manifests/kind-config.yaml
```

#### Install preprequisites & vault

Either use attached Makefile: `make kind_install`

Or do it manually:
1. Add repositories:
    ```shell
    helm repo add cert-manager https://charts.jetstack.io
    helm repo add kong https://charts.konghq.com
    helm repo add hashicorp https://helm.releases.hashicorp.com/
    helm repo update
    ```
2. Install cert-manager:
   ```shell
   helm upgrade --install cert-manager cert-manager/cert-manager \
     --namespace cert-manager \
     --create-namespace \
     --set installCRDs=true \
     --wait
   ```
3. Install kong ingress controller:
```shell
helm upgrade --install kong kong/kong \
  --namespace kong --create-namespace \
  --values manifests/kong-values.yml \
  --wait
```
4. Install vault:
```shell
kubectl create namespace vault || echo 0
kubectl apply -f manifests/certs.yml
helm upgrade --install --namespace vault \
  vault hashicorp/vault \
  --values manifests/vault-values.yml \
  --wait
```

#### Install vault-autounseal:
Either with makefile: `make helm_install` or use:
```shell
helm upgrade --install vault-autounseal-operator charts/vault-autounseal-operator -n vault
```
