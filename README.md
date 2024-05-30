# vault-autounseal-operator
Vault operator for managing vault clusters running in Kubernetes. Operator handles:
* automated initialization of new clusters
* automated unsealing of pods in a clusters
* upgrading statefulset pods in graceful manner
* rotating vault pods if TLS certificate is updated

Operator assumes vault is deployed using official Hashicorp Vault helm chart.

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