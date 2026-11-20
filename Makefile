IMG ?= postgres-provisioner:0.1.0
PLATFORMS ?= linux/amd64,linux/arm64

.PHONY: test build docker-buildx manifests generate helm-lint

test:
	go test ./...

build:
	CGO_ENABLED=0 go build -trimpath -o bin/manager ./cmd

docker-buildx:
	docker buildx build --platform $(PLATFORMS) -t $(IMG) --push .

manifests:
	go run sigs.k8s.io/controller-tools/cmd/controller-gen@v0.17.3 crd rbac:roleName=postgres-provisioner webhook paths="./..."

generate:
	go run sigs.k8s.io/controller-tools/cmd/controller-gen@v0.17.3 object:headerFile="" paths="./..."
