# syntax = docker/dockerfile:latest
# check=error=true
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

FROM ghcr.io/kevindiu/monorepo-go-example-buildbase:latest

LABEL maintainer="monorepo-go-example team"

# skipcq: DOK-DL3002
USER root:root

ARG TARGETARCH
ARG TARGETOS
ARG GO_VERSION=1.21.0
ARG GOLANGCI_LINT_VERSION=v1.55.2
ARG PROTOC_VERSION=24.4
ARG BUF_VERSION=1.28.1
ARG KUBECTL_VERSION=1.28.4
ARG DOCKER_VERSION=v24.0.7

ENV APP_NAME=ci-container
ENV DEBIAN_FRONTEND=noninteractive
ENV GO111MODULE=on
ENV GOPATH=/go
ENV GOROOT=/opt/go
ENV HOME=/root
ENV LANG=en_US.UTF-8
ENV LANGUAGE=en_US.UTF-8
ENV LC_ALL=en_US.UTF-8
ENV TZ=Etc/UTC
ENV USER=root
ENV PATH=${GOPATH}/bin:${GOROOT}/bin:/usr/local/bin:${PATH}

WORKDIR ${GOPATH}

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

#skipcq: DOK-W1001, DOK-SC2046, DOK-SC2086, DOK-DL3008
RUN --mount=type=cache,target=/var/lib/apt,sharing=locked \
    --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target="${GOPATH}/pkg",id="go-pkg-${TARGETARCH}" \
    --mount=type=cache,target="${HOME}/.cache/go-build",id="go-build-${TARGETARCH}" \
    set -ex \
    && echo 'Binary::apt::APT::Keep-Downloaded-Packages "true";' > /etc/apt/apt.conf.d/keep-cache \
    && echo 'APT::Install-Recommends "false";' > /etc/apt/apt.conf.d/no-install-recommends \
    && apt-get clean \
    && apt-get update -y \
    && apt-get upgrade -y \
    && apt-get install -y --no-install-recommends --fix-missing \
    sudo \
    docker.io \
    && ldconfig

# Install golangci-lint
RUN set -ex \
    && curl -fsSL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${GOPATH}/bin ${GOLANGCI_LINT_VERSION}

# Install protoc
RUN set -ex \
    && curl -fsSL https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-$(uname -m | sed 's/aarch64/aarch_64/').zip -o /tmp/protoc.zip \
    && unzip -q /tmp/protoc.zip -d /usr/local \
    && rm -f /tmp/protoc.zip

# Install buf
RUN set -ex \
    && curl -fsSL https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-Linux-$(uname -m).tar.gz -o /tmp/buf.tar.gz \
    && tar -xzf /tmp/buf.tar.gz -C /usr/local --strip-components=1 buf/bin/buf \
    && rm -f /tmp/buf.tar.gz

# Install kubectl
RUN set -ex \
    && curl -fsSL https://dl.k8s.io/release/v${KUBECTL_VERSION}/bin/linux/${TARGETARCH}/kubectl -o /usr/local/bin/kubectl \
    && chmod +x /usr/local/bin/kubectl

# Install Docker CLI
RUN set -ex \
    && curl -fsSL https://download.docker.com/linux/static/stable/$(uname -m)/docker-${DOCKER_VERSION#v}.tgz -o /tmp/docker.tgz \
    && tar -xzf /tmp/docker.tgz -C /tmp \
    && mv /tmp/docker/docker /usr/local/bin/ \
    && rm -rf /tmp/docker.tgz /tmp/docker

# Install Go tools
RUN set -ex \
    && go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest \
    && go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest \
    && go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

# skipcq: DOK-DL3002
USER root:root

ENTRYPOINT ["/bin/bash"]
