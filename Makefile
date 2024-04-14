all: clean test build

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

