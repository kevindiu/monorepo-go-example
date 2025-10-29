#
# Copyright (C) 2025 Kevin Diu <kevindiujp@gmail.com>
#
# Licensed under the Apache License, Version 2.0 (the "License");
# You may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

.DEFAULT_GOAL := help

# Project variables
REPO = github.com/kevindiu/monorepo-go-example
PROJECT_NAME = monorepo-go-example
VERSION ?= v0.1.0
MAINTAINER = kevindiu
GOPKG = $(REPO)

# Build variables
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
CGO_ENABLED ?= 1
GO_VERSION ?= $(shell go version | awk '{print $$3}' | sed 's/go//')

# Directory paths
ROOTDIR = $(shell pwd)
CMDDIR = $(ROOTDIR)/cmd
PKGDIR = $(ROOTDIR)/pkg
INTERNALDIR = $(ROOTDIR)/internal
BINDIR = $(ROOTDIR)/bin
BUILDDIR = $(ROOTDIR)/build

# Services
SERVICES = user-service order-service gateway

# Tools
TOOLS_DIR = $(ROOTDIR)/hack/tools
BUF_VERSION = v1.28.1
PROTOC_GEN_GO_VERSION = v1.31.0
PROTOC_GEN_GO_GRPC_VERSION = v1.3.0
PROTOC_GEN_GRPC_GATEWAY_VERSION = v2.18.1

# Docker/CI variables
ORG = kevindiu
GHCRORG = ghcr.io/$(ORG)
BUILDBASE_IMAGE = $(PROJECT_NAME)-buildbase
CI_CONTAINER_IMAGE = $(PROJECT_NAME)-ci-container
USER_SERVICE_IMAGE = $(PROJECT_NAME)-user-service
CONFIG_SERVICE_IMAGE = $(PROJECT_NAME)-config-service

# Tool versions from versions directory
TOOL_GO_VERSION = $(shell cat versions/GO_VERSION 2>/dev/null || echo "1.21.0")
GOLANGCI_LINT_VERSION = $(shell cat versions/GOLANGCI_LINT_VERSION 2>/dev/null || echo "v1.55.2")
PROTOC_VERSION = $(shell cat versions/PROTOC_VERSION 2>/dev/null || echo "24.4")
BUF_TOOL_VERSION = $(shell cat versions/BUF_VERSION 2>/dev/null || echo "1.28.1")
KUBECTL_VERSION = $(shell cat versions/KUBECTL_VERSION 2>/dev/null || echo "1.28.4")
DOCKER_VERSION = $(shell cat versions/DOCKER_VERSION 2>/dev/null || echo "v24.0.7")

# Build metadata
GIT_COMMIT = $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
TIMESTAMP = $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
TAG ?= latest
PLATFORM = linux/amd64,linux/arm64
BUILDKIT_INLINE_CACHE = 1
REMOTE ?= false
DOCKER = docker
ARCH = $(shell uname -m)
DOCKER_OPTS ?=
EXTRA_ARGS ?=

# Colors for output
RED = \033[0;31m
GREEN = \033[0;32m
YELLOW = \033[0;33m
BLUE = \033[0;34m
NC = \033[0m # No Color

.PHONY: help
## Show help
help:
	@echo '$(YELLOW)Available targets:$(NC)'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "  $(GREEN)%-20s$(NC) %s\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

$(BINDIR):
	mkdir -p $(BINDIR)

$(BUILDDIR):
	mkdir -p $(BUILDDIR)

.PHONY: deps
## Install dependencies
deps:
	@echo '$(BLUE)Installing dependencies...$(NC)'
	go mod download
	go mod tidy

.PHONY: tools
## Install development tools
tools:
	@echo '$(BLUE)Installing development tools...$(NC)'
	go install github.com/bufbuild/buf/cmd/buf@$(BUF_VERSION)
	go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@$(PROTOC_GEN_GRPC_GATEWAY_VERSION)
	go install github.com/google/uuid@latest

.PHONY: proto
## Generate protobuf code
proto:
	@echo '$(BLUE)Generating protobuf code...$(NC)'
	buf generate

.PHONY: build
## Build all services
build: $(BINDIR) proto
	@echo '$(BLUE)Building all services...$(NC)'
	@for service in $(SERVICES); do \
		echo "Building $$service..."; \
		CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build -o $(BINDIR)/$$service $(CMDDIR)/$$service; \
	done

.PHONY: build-user-service
## Build user service
build-user-service: $(BINDIR) proto
	@echo '$(BLUE)Building user service...$(NC)'
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) \
	go build -o $(BINDIR)/user-service $(CMDDIR)/user-service

.PHONY: build-order-service
## Build order service
build-order-service: $(BINDIR) proto
	@echo '$(BLUE)Building order service...$(NC)'
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) \
	go build -o $(BINDIR)/order-service $(CMDDIR)/order-service

.PHONY: build-gateway
## Build gateway service
build-gateway: $(BINDIR) proto
	@echo '$(BLUE)Building gateway service...$(NC)'
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) \
	go build -o $(BINDIR)/gateway $(CMDDIR)/gateway

.PHONY: run-user-service
## Run user service
run-user-service: build-user-service
	@echo '$(BLUE)Running user service...$(NC)'
	$(BINDIR)/user-service

.PHONY: run-order-service
## Run order service
run-order-service: build-order-service
	@echo '$(BLUE)Running order service...$(NC)'
	$(BINDIR)/order-service

.PHONY: run-gateway
## Run gateway service
run-gateway: build-gateway
	@echo '$(BLUE)Running gateway service...$(NC)'
	$(BINDIR)/gateway

.PHONY: test
## Run tests
test:
	@echo '$(BLUE)Running tests...$(NC)'
	go test -v -race -cover ./...

.PHONY: test-unit
## Run unit tests
test-unit:
	@echo '$(BLUE)Running unit tests...$(NC)'
	go test -v -race -cover -short ./internal/... ./pkg/*/service/...

.PHONY: test-integration
## Run integration tests
test-integration:
	@echo '$(BLUE)Running integration tests...$(NC)'
	go test -v -race -cover -tags=integration ./pkg/*/repository/...

.PHONY: test-e2e
## Run end-to-end tests
test-e2e:
	@echo '$(BLUE)Running E2E tests...$(NC)'
	go test -v -race -timeout 60s ./tests/e2e/...

.PHONY: test-coverage
## Run tests with coverage report
test-coverage:
	@echo '$(BLUE)Running tests with coverage...$(NC)'
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo '$(GREEN)Coverage report generated: coverage.html$(NC)'

.PHONY: lint
## Run linter
lint:
	@echo '$(BLUE)Running linter...$(NC)'
	golangci-lint run

.PHONY: fmt
## Format code
fmt:
	@echo '$(BLUE)Formatting code...$(NC)'
	go fmt ./...
	goimports -w .

.PHONY: vet
## Run go vet
vet:
	@echo '$(BLUE)Running go vet...$(NC)'
	go vet ./...

.PHONY: mod
## Update go modules
mod:
	@echo '$(BLUE)Updating go modules...$(NC)'
	go mod tidy
	go mod verify

.PHONY: clean
## Clean build artifacts
clean:
	@echo '$(BLUE)Cleaning build artifacts...$(NC)'
	rm -rf $(BINDIR)
	rm -rf $(BUILDDIR)
	go clean -cache
	go clean -testcache

.PHONY: docker-build
## Build Docker images for all services
docker-build:
	@echo '$(BLUE)Building Docker images...$(NC)'
	@for service in $(SERVICES); do \
		echo "Building Docker image for $$service..."; \
		docker build -f deployments/docker/$$service/Dockerfile -t $(PROJECT_NAME)/$$service:$(VERSION) .; \
	done

.PHONY: docker-run
## Run services with Docker Compose
docker-run:
	@echo '$(BLUE)Running services with Docker Compose...$(NC)'
	docker-compose -f deployments/docker-compose.yml up

.PHONY: docker-stop
## Stop Docker Compose services
docker-stop:
	@echo '$(BLUE)Stopping Docker Compose services...$(NC)'
	docker-compose -f deployments/docker-compose.yml down

.PHONY: k8s-deploy
## Deploy to Kubernetes
k8s-deploy:
	@echo '$(BLUE)Deploying to Kubernetes...$(NC)'
	kubectl apply -f deployments/k8s/

.PHONY: k8s-undeploy
## Remove from Kubernetes
k8s-undeploy:
	@echo '$(BLUE)Removing from Kubernetes...$(NC)'
	kubectl delete -f deployments/k8s/

.PHONY: helm-install
## Install Helm chart
helm-install:
	@echo '$(BLUE)Installing Helm chart...$(NC)'
	helm install $(PROJECT_NAME) charts/$(PROJECT_NAME)

.PHONY: helm-upgrade
## Upgrade Helm chart
helm-upgrade:
	@echo '$(BLUE)Upgrading Helm chart...$(NC)'
	helm upgrade $(PROJECT_NAME) charts/$(PROJECT_NAME)

.PHONY: helm-uninstall
## Uninstall Helm chart
helm-uninstall:
	@echo '$(BLUE)Uninstalling Helm chart...$(NC)'
	helm uninstall $(PROJECT_NAME)

.PHONY: dev-setup
## Setup development environment
dev-setup: deps tools proto
	@echo '$(GREEN)Development environment setup complete!$(NC)'

.PHONY: all
## Build everything
all: deps proto build test

.PHONY: ci
## Run CI pipeline
ci: deps proto build test-unit lint vet

.PHONY: release
## Create release builds
release: clean proto
	@echo '$(BLUE)Creating release builds...$(NC)'
	@for service in $(SERVICES); do \
		for os in linux darwin windows; do \
			for arch in amd64 arm64; do \
				if [ "$$os" = "windows" ]; then \
					ext=".exe"; \
				else \
					ext=""; \
				fi; \
				echo "Building $$service for $$os/$$arch..."; \
				CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch \
				go build -ldflags="-s -w" -o $(BUILDDIR)/$$service-$$os-$$arch$$ext $(CMDDIR)/$$service; \
			done; \
		done; \
	done

.PHONY: version
## Show version information
version:
	@echo "Version: $(VERSION)"
	@echo "Go Version: $(GO_VERSION)"
	@echo "OS/Arch: $(GOOS)/$(GOARCH)"

# Include docker targets
-include Makefile.d/docker.mk

.PHONY: docker/platforms
## print docker platforms
docker/platforms:
	@echo "linux/amd64,linux/arm64"

.PHONY: docker/darch
docker/darch:
	@echo $(subst x86_64,amd64,$(subst aarch64,arm64,$(ARCH)))

.PHONY: docker/platform
docker/platform:
	@echo linux/$(shell $(MAKE) -s docker/darch)

.PHONY: docker/build/all
## Build all docker images
docker/build/all: \
	docker/build/buildbase \
	docker/build/ci-container \
	docker/build/user-service \
	docker/build/config-service

.PHONY: docker/build/image
## Generalized docker build function
docker/build/image:
ifeq ($(REMOTE),true)
	@echo "starting remote build for $(IMAGE):$(TAG)"
	DOCKER_BUILDKIT=1 $(DOCKER) buildx build \
		$(DOCKER_OPTS) \
		--cache-to type=gha,scope=$(TAG)-buildcache,mode=max \
		--cache-from type=gha,scope=$(TAG)-buildcache \
		--build-arg BUILDKIT_INLINE_CACHE=$(BUILDKIT_INLINE_CACHE) \
		--build-arg GO_VERSION=$(TOOL_GO_VERSION) \
		--build-arg GOLANGCI_LINT_VERSION=$(GOLANGCI_LINT_VERSION) \
		--build-arg PROTOC_VERSION=$(PROTOC_VERSION) \
		--build-arg BUF_VERSION=$(BUF_TOOL_VERSION) \
		--build-arg KUBECTL_VERSION=$(KUBECTL_VERSION) \
		--build-arg DOCKER_VERSION=$(DOCKER_VERSION) \
		$(EXTRA_ARGS) \
		--label org.opencontainers.image.url=$(REPO) \
		--label org.opencontainers.image.source=https://$(REPO) \
		--label org.opencontainers.image.vendor=$(ORG) \
		--label org.opencontainers.image.version=$(VERSION) \
		--label org.opencontainers.image.created=$(TIMESTAMP) \
		--label org.opencontainers.image.revision=$(GIT_COMMIT) \
		--platform $(PLATFORM) \
		--pull \
		--file $(DOCKERFILE) \
		--tag $(ORG)/$(IMAGE):$(TAG) \
		--tag $(GHCRORG)/$(IMAGE):$(TAG) \
		$(ROOTDIR)
else
	@echo "starting local build for $(IMAGE):$(TAG)"
	DOCKER_BUILDKIT=1 $(DOCKER) build \
		$(DOCKER_OPTS) \
		--build-arg BUILDKIT_INLINE_CACHE=$(BUILDKIT_INLINE_CACHE) \
		--build-arg GO_VERSION=$(TOOL_GO_VERSION) \
		--build-arg GOLANGCI_LINT_VERSION=$(GOLANGCI_LINT_VERSION) \
		--build-arg PROTOC_VERSION=$(PROTOC_VERSION) \
		--build-arg BUF_VERSION=$(BUF_TOOL_VERSION) \
		--build-arg KUBECTL_VERSION=$(KUBECTL_VERSION) \
		--build-arg DOCKER_VERSION=$(DOCKER_VERSION) \
		$(EXTRA_ARGS) \
		--label org.opencontainers.image.url=$(REPO) \
		--label org.opencontainers.image.source=https://$(REPO) \
		--label org.opencontainers.image.vendor=$(ORG) \
		--label org.opencontainers.image.version=$(VERSION) \
		--label org.opencontainers.image.created=$(TIMESTAMP) \
		--label org.opencontainers.image.revision=$(GIT_COMMIT) \
		--platform $(shell $(MAKE) -s docker/platform) \
		--file $(DOCKERFILE) \
		--tag $(ORG)/$(IMAGE):$(TAG) \
		--tag $(GHCRORG)/$(IMAGE):$(TAG) \
		$(ROOTDIR)
endif

.PHONY: docker/name/buildbase
## print buildbase image name
docker/name/buildbase:
	@echo "$(ORG)/$(BUILDBASE_IMAGE)"

.PHONY: docker/build/buildbase
## build buildbase image
docker/build/buildbase:
	@make DOCKERFILE="$(ROOTDIR)/dockers/buildbase/Dockerfile" \
		IMAGE=$(BUILDBASE_IMAGE) \
		docker/build/image

.PHONY: docker/name/ci-container
## print ci-container image name
docker/name/ci-container:
	@echo "$(ORG)/$(CI_CONTAINER_IMAGE)"

.PHONY: docker/build/ci-container
## build ci-container image
docker/build/ci-container:
	@make DOCKERFILE="$(ROOTDIR)/dockers/ci/base/Dockerfile" \
		IMAGE=$(CI_CONTAINER_IMAGE) \
		docker/build/image

.PHONY: docker/name/user-service
## print user-service image name
docker/name/user-service:
	@echo "$(ORG)/$(USER_SERVICE_IMAGE)"

.PHONY: docker/build/user-service
## build user-service image
docker/build/user-service:
	@make DOCKERFILE="$(ROOTDIR)/cmd/user-service/Dockerfile" \
		IMAGE=$(USER_SERVICE_IMAGE) \
		docker/build/image

.PHONY: docker/name/config-service
## print config-service image name
docker/name/config-service:
	@echo "$(ORG)/$(CONFIG_SERVICE_IMAGE)"

.PHONY: docker/build/config-service
## build config-service image
docker/build/config-service:
	@make DOCKERFILE="$(ROOTDIR)/cmd/config-service/Dockerfile" \
		IMAGE=$(CONFIG_SERVICE_IMAGE) \
		docker/build/image

.PHONY: docker/push/all
## push all docker images
docker/push/all: \
	docker/push/buildbase \
	docker/push/ci-container \
	docker/push/user-service \
	docker/push/config-service

.PHONY: docker/push/%
## push specific docker image
docker/push/%:
	$(eval IMAGE_NAME := $(shell make -s docker/name/$*))
	$(DOCKER) push $(IMAGE_NAME):$(TAG)
	$(DOCKER) push $(GHCRORG)/$(IMAGE_NAME):$(TAG)
