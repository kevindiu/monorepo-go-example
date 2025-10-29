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

# skipcq: DOK-DL3007
FROM ubuntu:latest AS builder

LABEL maintainer="monorepo-go-example team"

# skipcq: DOK-DL3002
USER root:root

ARG TARGETARCH
ARG TARGETOS
ARG GO_VERSION=1.21.0

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
    set -ex \
    && echo 'Binary::apt::APT::Keep-Downloaded-Packages "true";' > /etc/apt/apt.conf.d/keep-cache \
    && echo 'APT::Install-Recommends "false";' > /etc/apt/apt.conf.d/no-install-recommends \
    && apt-get clean \
    && apt-get update -y \
    && apt-get upgrade -y \
    && apt-get install -y --no-install-recommends --fix-missing \
    build-essential \
    ca-certificates \
    curl \
    tzdata \
    locales \
    git \
    cmake \
    g++ \
    gcc \
    libssl-dev \
    unzip \
    pkg-config \
    file \
    gawk \
    gnupg2 \
    jq \
    sed \
    zip \
    && ldconfig \
    && locale-gen en_US.UTF-8 \
    && update-locale LANG=en_US.UTF-8 LANGUAGE=en_US.UTF-8 LC_ALL=en_US.UTF-8 \
    && ln -snf /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo ${TZ} > /etc/timezone

# Install Go
RUN set -ex \
    && curl -fsSL https://go.dev/dl/go${GO_VERSION}.linux-${TARGETARCH}.tar.gz -o /tmp/go.tar.gz \
    && tar -C /opt -xzf /tmp/go.tar.gz \
    && rm -f /tmp/go.tar.gz

# skipcq: DOK-DL3002
USER root:root

ENTRYPOINT ["/bin/bash"]
