#
# Copyright (C) 2024 monorepo-go-example
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

.PHONY: docker/build
## build all docker images
docker/build: \
	docker/build/buildbase \
	docker/build/ci-container \
	docker/build/user-service \
	docker/build/config-service

.PHONY: docker/name/org
## print docker organization name
docker/name/org:
	@echo "$(ORG)"

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
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg GOLANGCI_LINT_VERSION=$(GOLANGCI_LINT_VERSION) \
		--build-arg PROTOC_VERSION=$(PROTOC_VERSION) \
		--build-arg BUF_VERSION=$(BUF_VERSION) \
		--build-arg KUBECTL_VERSION=$(KUBECTL_VERSION) \
		--build-arg DOCKER_VERSION=$(DOCKER_VERSION) \
		$(EXTRA_ARGS) \
		--label org.opencontainers.image.url=$(REPO_URL) \
		--label org.opencontainers.image.source=$(REPO_URL) \
		--label org.opencontainers.image.vendor=$(VENDOR) \
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
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg GOLANGCI_LINT_VERSION=$(GOLANGCI_LINT_VERSION) \
		--build-arg PROTOC_VERSION=$(PROTOC_VERSION) \
		--build-arg BUF_VERSION=$(BUF_VERSION) \
		--build-arg KUBECTL_VERSION=$(KUBECTL_VERSION) \
		--build-arg DOCKER_VERSION=$(DOCKER_VERSION) \
		$(EXTRA_ARGS) \
		--label org.opencontainers.image.url=$(REPO_URL) \
		--label org.opencontainers.image.source=$(REPO_URL) \
		--label org.opencontainers.image.vendor=$(VENDOR) \
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

.PHONY: docker/push
## push all docker images
docker/push: \
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
