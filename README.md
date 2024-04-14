# vault-autounseal-operator
Vault operator for managing vault clusters running in Kubernetes. Operator handles:
* automated initialization of new clusters
* automated unsealing of pods in a clusters
* upgrading statefulset pods in graceful manner
* rotating vault pods if TLS certificate is updated

Operator assumes vault is deployed using official Hashicorp Vault helm chart.
