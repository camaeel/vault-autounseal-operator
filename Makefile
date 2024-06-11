all: clean test build


release:
	goreleaser release --snapshot --clean

generate: 
	go generate ./...

test: generate
	go test ./... -v

clean:
	rm -rf bin || true

build: generate
	go build -o bin/vault-autounseal-operator github.com/camaeel/vault-autounseal-operator/cmd/vault-autounseal-operator

docker:
	docker buildx build -t vault-autounseal-operator:local --load .

docker_debug:
	docker buildx build -t vault-autounseal-operator:debug --target=debug --build-arg DEBUG=1 --load .

# autounseal_kind: docker docker_kind_load
# 	kubectl run --rm -it --image vault-autounseal-operator:local test --command -- /vault-autounseal

docker_kind_load: docker
	kind load docker-image vault-autounseal-operator:local

docker_debug_kind_load: docker_debug
	kind load docker-image vault-autounseal-operator:debug

kind:
	kind create cluster \
		--wait 120s \
		--config manifests/kind-config.yaml

kind_install:
	helm repo add cert-manager https://charts.jetstack.io
	helm repo update
	helm upgrade --install cert-manager cert-manager/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --set installCRDs=true \
  --wait
	helm repo add kong https://charts.konghq.com 
	helm repo update
	helm upgrade --install kong kong/kong \
		--namespace kong --create-namespace \
		--values manifests/kong-values.yml \
		--wait
	kubectl create namespace vault || echo 0
	kubectl apply -f manifests/certs.yml
	helm repo add hashicorp https://helm.releases.hashicorp.com/
	helm repo update
	helm upgrade --install --namespace vault \
		vault hashicorp/vault \
		--values manifests/vault-values.yml \
		--wait

helm_template:
	helm template vault-autounseal-operator charts/vault-autounseal-operator -n vault --debug
helm_install:
	helm upgrade --install vault-autounseal-operator charts/vault-autounseal-operator -n vault
