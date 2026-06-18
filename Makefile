# Demo: HTTP/2 persistent-connection load balancing — the problem (no-mesh) and
# the service-mesh solution (mesh). Everything runs in containers; the host needs
# only Docker. CLI tooling is downloaded into ./bin (no apt/snap), and the Go
# build happens inside Docker (no Go toolchain on the host).
#
# Usual flow:
#   make dev           # build + cluster + Istio + apply both namespaces
#   make demo-no-mesh  # watch the SAME UUID repeat (Arco 4)
#   make demo-mesh     # watch 3 UUIDs alternate ~60/20/20 (Arco 7)
#   make clean         # tear everything down, leave no trace on the host

SHELL := /bin/bash
ROOT  := $(shell pwd)
BIN   := $(ROOT)/bin

include versions.env

# Use the vendored tools and a project-local kubeconfig so we never touch the
# host PATH or ~/.kube/config.
export PATH       := $(BIN):$(PATH)
export KUBECONFIG := $(BIN)/kubeconfig

# Local Kubernetes runtime. Only k3d is wired up; documented here for future swap.
CLUSTER_TOOL ?= k3d

OS   := $(shell uname -s | tr '[:upper:]' '[:lower:]')
RAW_ARCH := $(shell uname -m)
ARCH := $(if $(filter x86_64,$(RAW_ARCH)),amd64,$(if $(filter aarch64 arm64,$(RAW_ARCH)),arm64,$(RAW_ARCH)))

# go-api Deployments to wait on.
API_DEPLOYS := go-api-v1 go-api-v2 go-api-v3

.PHONY: help bootstrap go.sum build cluster-up istio-install images-import \
        dev demo-no-mesh demo-mesh verify down clean

help: ## Show this help.
	@grep -E '^[a-zA-Z0-9_.-]+:.*## ' $(MAKEFILE_LIST) | sort | \
	  awk 'BEGIN{FS=":.*## "}{printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

## ---- tooling (downloaded into ./bin) -------------------------------------

bootstrap: $(BIN)/k3d $(BIN)/kubectl $(BIN)/istioctl ## Download k3d, kubectl, istioctl into ./bin.

$(BIN)/k3d:
	@mkdir -p $(BIN)
	@echo ">> downloading k3d $(K3D_VERSION)"
	curl -sSfL https://github.com/k3d-io/k3d/releases/download/$(K3D_VERSION)/k3d-$(OS)-$(ARCH) -o $(BIN)/k3d
	chmod +x $(BIN)/k3d

$(BIN)/kubectl:
	@mkdir -p $(BIN)
	@echo ">> downloading kubectl $(KUBECTL_VERSION)"
	curl -sSfL https://dl.k8s.io/release/$(KUBECTL_VERSION)/bin/$(OS)/$(ARCH)/kubectl -o $(BIN)/kubectl
	chmod +x $(BIN)/kubectl

$(BIN)/istioctl:
	@mkdir -p $(BIN)
	@echo ">> downloading istioctl $(ISTIO_VERSION)"
	curl -sSfL https://github.com/istio/istio/releases/download/$(ISTIO_VERSION)/istioctl-$(ISTIO_VERSION)-$(OS)-$(ARCH).tar.gz | tar -xz -C $(BIN) istioctl
	chmod +x $(BIN)/istioctl

## ---- build (inside Docker) -----------------------------------------------

go.sum: go.mod ## Resolve Go deps (runs `go mod tidy` in a container).
	docker run --rm -u $$(id -u):$$(id -g) \
	  -e HOME=/tmp -e GOCACHE=/tmp/.cache -e GOPATH=/tmp/go -e GOFLAGS=-mod=mod \
	  -v $(ROOT):/src -w /src golang:$(GO_VERSION) go mod tidy

build: go.sum ## Build the server and client images via Docker.
	# Offline-safe after the first online run: no `--pull`, so cached base images are
	# reused, and the `go mod download` layer is keyed by go.mod/go.sum (unchanged code
	# rebuilds hit the cache). A repeat `make dev` offline does no network I/O here.
	docker build -f Dockerfile.server --build-arg GO_VERSION=$(GO_VERSION) -t $(SERVER_IMAGE) .
	docker build -f Dockerfile.client --build-arg GO_VERSION=$(GO_VERSION) -t $(CLIENT_IMAGE) .

## ---- cluster + istio ------------------------------------------------------

cluster-up: bootstrap ## Create the k3d cluster (idempotent) and write ./bin/kubeconfig.
	@if ! k3d cluster list $(CLUSTER_NAME) >/dev/null 2>&1; then \
	  echo ">> creating k3d cluster $(CLUSTER_NAME)"; \
	  k3d cluster create $(CLUSTER_NAME) --image $(K3S_IMAGE) --servers 1 --agents 2 \
	    --k3s-arg "--disable=traefik@server:*" \
	    --kubeconfig-update-default=false --kubeconfig-switch-context=false --wait; \
	else \
	  echo ">> k3d cluster $(CLUSTER_NAME) already exists"; \
	fi
	@k3d kubeconfig get $(CLUSTER_NAME) > $(KUBECONFIG)

istio-install: cluster-up ## Install Istio (minimal profile). Idempotent — skips (no network) if istiod already present.
	@if kubectl -n istio-system get deploy istiod >/dev/null 2>&1; then \
	  echo ">> istio already installed, skipping"; \
	else \
	  echo ">> installing istio $(ISTIO_VERSION)"; \
	  istioctl install --set profile=minimal -y; \
	fi

images-import: build cluster-up ## Load the local images into the cluster.
	k3d image import $(SERVER_IMAGE) $(CLIENT_IMAGE) -c $(CLUSTER_NAME)

## ---- the demo -------------------------------------------------------------

dev: istio-install images-import ## Bring everything up: cluster, Istio, both namespaces. Idempotent.
	@echo ">> ensuring namespaces"
	kubectl create namespace no-mesh --dry-run=client -o yaml | kubectl apply -f -
	kubectl create namespace mesh    --dry-run=client -o yaml | kubectl apply -f -
	kubectl label namespace mesh istio-injection=enabled --overwrite
	@echo ">> applying manifests"
	kubectl apply -k deploy/no-mesh
	kubectl apply -k deploy/mesh
	@echo ">> waiting for readiness"
	@for ns in no-mesh mesh; do \
	  for d in $(API_DEPLOYS) go-client; do \
	    kubectl -n $$ns rollout status deploy/$$d --timeout=120s; \
	  done; \
	done
	@echo ">> ready. Try: make demo-no-mesh  |  make demo-mesh"

demo-no-mesh: ## Stream the client logs in no-mesh (expect the SAME UUID repeating).
	kubectl -n no-mesh logs -f deploy/go-client

demo-mesh: ## Stream the client logs in mesh (expect 3 UUIDs alternating ~60/20/20).
	kubectl -n mesh logs -f deploy/go-client

verify: ## Sanity checks: pods 2/2 in mesh, Service port named http2.
	@echo "== no-mesh pods =="; kubectl -n no-mesh get pods
	@echo "== mesh pods (expect 2/2) =="; kubectl -n mesh get pods
	@echo "== mesh go-api Service port (expect name: http2) =="; \
	  kubectl -n mesh get svc go-api -o jsonpath='{.spec.ports[0].name}{"\n"}'

## ---- teardown -------------------------------------------------------------

down: ## Remove the workloads from both namespaces (keep the cluster).
	-kubectl delete -k deploy/mesh --ignore-not-found
	-kubectl delete -k deploy/no-mesh --ignore-not-found

clean: ## Delete the k3d cluster and the downloaded tooling (no trace on host).
	-$(BIN)/k3d cluster delete $(CLUSTER_NAME)
	rm -rf $(BIN)
