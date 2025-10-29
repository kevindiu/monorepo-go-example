# Monorepo Go Example CI Container

This image is designed for running CI workflows on GitHub Actions, following production-grade best practices.

## Overview

The CI container includes all tools needed for building, testing, and deploying the monorepo-go-example project.

<div align="center">
    <img src="https://img.shields.io/docker/v/kevindiu/monorepo-go-example-ci-container/latest?label=ci-container" alt="Latest Image"/>
    <img src="https://img.shields.io/badge/platforms-linux%2Famd64%20%7C%20linux%2Farm64-blue" alt="Platforms"/>
</div>

## Features

### Pre-installed Tools

- **Go** (`versions/GO_VERSION`): Go programming language
- **golangci-lint** (`versions/GOLANGCI_LINT_VERSION`): Comprehensive linter
- **protoc** (`versions/PROTOC_VERSION`): Protocol Buffers compiler
- **buf** (`versions/BUF_VERSION`): Modern Protobuf tooling
- **kubectl** (`versions/KUBECTL_VERSION`): Kubernetes CLI
- **Docker CLI** (`versions/DOCKER_VERSION`): Docker command-line client
- **protoc-gen-go**: Go code generator for protobuf
- **protoc-gen-go-grpc**: gRPC code generator
- **protoc-gen-grpc-gateway**: gRPC gateway generator
- **protoc-gen-openapiv2**: OpenAPI v2 generator

### System Packages

- `build-essential`: C/C++ compilers and build tools
- `git`: Version control
- `curl`: Data transfer tool
- `jq`: JSON processor
- `sudo`: Superuser privileges
- And more...

## Versions

| Tag | linux/amd64 | linux/arm64 | Description |
|-----|:-----------:|:-----------:|-------------|
| latest | ✅ | ✅ | Latest stable build from main branch |
| nightly | ✅ | ✅ | Daily automated build |
| vX.Y.Z | ✅ | ✅ | Specific version release |
| pr-XXX | ✅ | ❌ | Pull request preview build |

## Requirements

### linux/amd64
- CPU: x86_64 architecture
- Memory: Minimum 2GB RAM recommended

### linux/arm64
- CPU: ARM64 architecture (NOT Apple Silicon - use amd64 with Rosetta)
- Memory: Minimum 2GB RAM recommended

## Usage

### GitHub Actions

Use the CI container in your GitHub Actions workflows:

```yaml
name: CI Pipeline
on:
  push:
    branches:
      - main

jobs:
  test:
    name: Run Tests
    runs-on: ubuntu-latest
    container:
      image: ghcr.io/kevindiu/monorepo-go-example-ci-container:latest
      credentials:
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Run tests
        run: make test
```

### Local Development

Pull and run the container locally:

```bash
# Pull the latest image
docker pull ghcr.io/kevindiu/monorepo-go-example-ci-container:latest

# Run interactively
docker run -it --rm \
  -v $(pwd):/workspace \
  -w /workspace \
  ghcr.io/kevindiu/monorepo-go-example-ci-container:latest \
  /bin/bash

# Inside the container
make test
make lint
make build
```

### Docker Compose

Use in docker-compose.yml:

```yaml
version: '3.8'
services:
  ci:
    image: ghcr.io/kevindiu/monorepo-go-example-ci-container:latest
    volumes:
      - .:/workspace
    working_dir: /workspace
    command: make test
```

## Environment Variables

The following environment variables are pre-configured:

- `GO111MODULE=on`: Enable Go modules
- `GOPATH=/go`: Go workspace path
- `GOROOT=/opt/go`: Go installation path
- `PATH`: Includes Go binaries, cargo binaries, and local binaries
- `LANG`, `LANGUAGE`, `LC_ALL`: UTF-8 locale settings

## Building

### Using Make

```bash
# Build buildbase first
make docker/build/buildbase

# Build CI container
make docker/build/ci-container

# Build for specific platforms
make PLATFORM=linux/amd64 docker/build/ci-container
make PLATFORM=linux/arm64 docker/build/ci-container

# Build with remote caching
make REMOTE=true docker/build/ci-container
```

### Using Docker CLI

```bash
# Build locally
docker build -f dockers/ci/base/Dockerfile -t my-ci-container .

# Build with BuildKit
DOCKER_BUILDKIT=1 docker build \
  --build-arg GO_VERSION=1.21.0 \
  --build-arg GOLANGCI_LINT_VERSION=v1.55.2 \
  -f dockers/ci/base/Dockerfile \
  -t my-ci-container .
```

## Customization

### Updating Tool Versions

Edit version files in `versions/` directory:

```bash
echo "1.22.0" > versions/GO_VERSION
echo "v1.56.0" > versions/GOLANGCI_LINT_VERSION

# Rebuild
make docker/build/ci-container
```

### Adding New Tools

Modify `dockers/ci/base/Dockerfile`:

```dockerfile
# Install new tool
RUN set -ex \
    && curl -fsSL https://example.com/tool -o /usr/local/bin/tool \
    && chmod +x /usr/local/bin/tool
```

## Performance

### Build Time Comparison

| Stage | Without Cache | With Cache | Speedup |
|-------|--------------|------------|---------|
| Buildbase | ~5 minutes | ~30 seconds | 10x |
| CI Container | ~10 minutes | ~1 minute | 10x |

### Cache Strategy

The CI container uses multi-level caching:

1. **GitHub Actions Cache**: Shared across workflow runs
2. **Registry Cache**: Stored in GHCR
3. **BuildKit Cache**: Local build cache
4. **Go Module Cache**: Persistent Go dependency cache

## Architecture

```
┌─────────────────────────────────────────┐
│    CI Container                        │
│  ┌─────────────────────────────────┐   │
│  │ Build Tools                      │   │
│  │ - golangci-lint                  │   │
│  │ - protoc, buf                    │   │
│  │ - kubectl, docker                │   │
│  │ - Go tools                       │   │
│  └─────────────────────────────────┘   │
│              ↓ Built on                 │
│  ┌─────────────────────────────────┐   │
│  │ Buildbase                        │   │
│  │ - Ubuntu 24.04                   │   │
│  │ - Go 1.21.0                      │   │
│  │ - Basic build tools              │   │
│  └─────────────────────────────────┘   │
└─────────────────────────────────────────┘
```


## Feature Matrix

| Feature | Status |
|---------|--------|
| Multi-architecture | Complete |
| BuildKit caching | Complete |
| Tool version management | Complete |
| Pre-installed Go | Complete |
| Pre-installed protoc | Complete |
| Pre-installed kubectl | Complete |
| Advanced dependencies (NGT, FAISS) | Not needed |
| Kubernetes integration (kind, k3d) | Planned |
| Helm tools | Planned |

## Troubleshooting

### Permission Denied Errors

If you encounter permission errors in GitHub Actions:

```yaml
# Add correct permissions to job
permissions:
  contents: read
  packages: read  # Required to pull from GHCR
```

### Image Pull Failures

For private repositories, authenticate first:

```yaml
- name: Login to GHCR
  uses: docker/login-action@v3
  with:
    registry: ghcr.io
    username: ${{ github.actor }}
    password: ${{ secrets.GITHUB_TOKEN }}
```

### Out of Disk Space

The CI container is optimized to minimize size:

```bash
# Check image size
docker images | grep ci-container

# Prune unused images
docker image prune -af
```

## Development

### Local Testing

Test changes locally before committing:

```bash
# Build locally
make docker/build/ci-container

# Test in container
docker run -it --rm \
  -v $(pwd):/workspace \
  -w /workspace \
  kevindiu/monorepo-go-example-ci-container:latest \
  bash -c "make test && make lint"
```

### CI/CD Workflow

The CI container is built automatically:

- **Daily**: Nightly builds from main branch
- **On Push**: When Dockerfile or versions change
- **On Tag**: For version releases

## Contacts

- **Repository**: [github.com/kevindiu/monorepo-go-example](https://github.com/kevindiu/monorepo-go-example)
- **Issues**: [GitHub Issues](https://github.com/kevindiu/monorepo-go-example/issues)

## License

Copyright (C) 2024 monorepo-go-example

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.


