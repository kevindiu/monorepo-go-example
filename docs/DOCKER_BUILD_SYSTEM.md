# Docker Build System

This document describes the Docker build system for a production-grade CI/CD infrastructure.

## Architecture Overview

The build system consists of several layers:

```
┌─────────────────────────────────────────────────┐
│           Service Images                        │
│  (user-service, config-service, etc.)           │
└────────────────┬────────────────────────────────┘
                 │ FROM
┌────────────────┴────────────────────────────────┐
│          CI Container Image                     │
│  (All build tools: Go, protoc, buf, etc.)       │
└────────────────┬────────────────────────────────┘
                 │ FROM
┌────────────────┴────────────────────────────────┐
│          Buildbase Image                        │
│  (Basic build environment: Ubuntu + Go)         │
└─────────────────────────────────────────────────┘
```

## Key Features

### 1. Multi-Stage Base Images

**Buildbase Image** (`dockers/buildbase/Dockerfile`):
- Ubuntu-based foundation
- Pre-installed Go toolchain
- Basic build tools (gcc, make, git, etc.)
- Optimized for build caching

**CI Container Image** (`dockers/ci/base/Dockerfile`):
- Extends buildbase
- All CI/CD tools pre-installed:
  - `golangci-lint` - Linting
  - `protoc` - Protocol Buffers compiler
  - `buf` - Proto tooling
  - `kubectl` - Kubernetes CLI
  - Docker CLI
  - Go development tools

### 2. Version Management

All tool versions are centrally managed in the `versions/` directory:

```
versions/
├── GO_VERSION              # 1.21.0
├── GOLANGCI_LINT_VERSION   # v1.55.2
├── PROTOC_VERSION          # 24.4
├── BUF_VERSION             # 1.28.1
├── KUBECTL_VERSION         # 1.28.4
└── DOCKER_VERSION          # v24.0.7
```

**Benefits**:
- Single source of truth for versions
- Easy version updates across all images
- Reproducible builds

### 3. BuildKit Cache Mounts

The build system leverages BuildKit's advanced caching:

```dockerfile
RUN --mount=type=cache,target=/var/lib/apt,sharing=locked \
    --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target="${GOPATH}/pkg",id="go-pkg-${TARGETARCH}" \
    --mount=type=cache,target="${HOME}/.cache/go-build",id="go-build-${TARGETARCH}" \
    <build commands>
```

**Cache Types**:
- `apt` cache - Speeds up package installations
- `go-pkg` - Caches Go module downloads
- `go-build` - Caches Go build artifacts

### 4. Multi-Architecture Support

All images support both `linux/amd64` and `linux/arm64`:

```bash
# Build for multiple architectures
make PLATFORM=linux/amd64,linux/arm64 docker/build/ci-container
```

**Implementation**:
- QEMU for cross-compilation
- `TARGETARCH` ARG for architecture-specific logic
- Architecture-specific cache keys

## Makefile Targets

### Docker Build Targets

```bash
# Build all images
make docker/build/all

# Build specific images
make docker/build/buildbase
make docker/build/ci-container
make docker/build/user-service
make docker/build/config-service

# Print image names
make docker/name/buildbase
make docker/name/ci-container

# Push images
make docker/push/all
make docker/push/buildbase
```

### Docker Build Options

```bash
# Local build (default)
make docker/build/buildbase

# Remote build with GitHub Actions cache
make REMOTE=true docker/build/buildbase

# Custom tag
make TAG=v1.0.0 docker/build/buildbase

# Custom platform
make PLATFORM=linux/amd64 docker/build/buildbase
```

## GitHub Actions Workflows

### Workflow Structure

```
.github/workflows/
├── _docker-build.yml          # Reusable workflow for building images
├── docker-buildbase.yml       # Build buildbase image (weekly)
├── docker-ci-container.yml    # Build CI container (daily)
├── ci.yml                     # Main CI pipeline using CI container
├── pr-validation.yml          # Fast PR checks
├── e2e-tests.yml              # E2E tests
└── release.yml                # Release automation
```

### Reusable Workflow Pattern

We use a reusable workflow (`_docker-build.yml`) that all image builds call:

```yaml
jobs:
  build:
    uses: "./.github/workflows/_docker-build.yml"
    with:
      target: "ci-container"
      platforms: "linux/amd64,linux/arm64"
    permissions:
      contents: read
      packages: write
    secrets:
      GHCR_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Benefits**:
- DRY (Don't Repeat Yourself)
- Consistent build process
- Easy to maintain and update

### CI Container Usage

The enhanced CI pipeline uses the pre-built CI container for all jobs:

```yaml
jobs:
  unit-tests:
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
        run: make test-unit
```

**Advantages**:
- Fast CI runs (no tool installation time)
- Consistent environment (same tools locally and in CI)
- Reduced workflow complexity

## Build Process

### 1. Buildbase Image Build

```bash
# Triggered: Weekly, on main push, or version changes
make docker/build/buildbase
```

**Process**:
1. Pull Ubuntu latest
2. Install system packages (gcc, make, git, etc.)
3. Install Go from version file
4. Configure environment variables
5. Push to registries (Docker Hub, GHCR)

### 2. CI Container Build

```bash
# Triggered: Daily, on main push, or Dockerfile changes
make docker/build/ci-container
```

**Process**:
1. Pull buildbase image
2. Install CI/CD tools:
   - golangci-lint
   - protoc
   - buf
   - kubectl
   - Docker CLI
   - Go tools (protoc-gen-go, etc.)
3. Configure additional environments
4. Push to registries

### 3. Service Image Build

```bash
# Triggered: On PR, main push, or tags
make docker/build/user-service
```

**Process**:
1. Pull buildbase image
2. Build service binary with caching
3. Create minimal runtime image (distroless)
4. Multi-stage build optimization
5. Push to registries

## Feature Matrix

| Feature | Status |
|---------|--------|
| Base image hierarchy | Complete |
| Version management | Complete |
| BuildKit cache mounts | Complete |
| Multi-arch builds | Complete |
| Reusable workflows | Complete |
| CI container pattern | Complete |
| Makefile organization | Complete |

## Advanced Features

### 1. GitHub Actions Cache

Both local and remote caching:

```yaml
cache-from: |
  type=gha,scope=${{ inputs.target }}-buildcache
  type=registry,ref=ghcr.io/...
cache-to: |
  type=gha,scope=${{ inputs.target }}-buildcache,mode=max
  type=registry,ref=ghcr.io/...,mode=max
```

### 2. Permissions Model

Following least-privilege principle:

```yaml
permissions:
  contents: read    # Read repository
  packages: write   # Push to GHCR
```

### 3. Secrets Management

Secure handling of credentials:

```yaml
secrets:
  DOCKERHUB_USER: ${{ secrets.DOCKERHUB_USER }}
  DOCKERHUB_PASS: ${{ secrets.DOCKERHUB_PASS }}
  GHCR_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Future Enhancements

Planned features:

### 1. Dockerfile Generation

Create `hack/docker/gen/main.go` to auto-generate Dockerfiles:

```go
// Generate Dockerfiles from templates
// Benefits:
// - DRY across similar images
// - Automated dependency tree analysis
// - Consistent structure
```

### 2. Workflow Generation

Auto-generate GitHub Actions workflows:

```go
// Generate workflows from configuration
// Benefits:
// - Consistency across workflows
// - Easy to add new services
// - Reduced boilerplate
```

### 3. Dependency Management

Automated tool updates:

```bash
make update/go        # Update Go version
make update/protoc    # Update protoc version
make update/buf       # Update buf version
```

## Best Practices

### 1. Cache Optimization

```dockerfile
# ✅ Good: Separate layers for better caching
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build

# ❌ Bad: Single layer defeats caching
COPY . .
RUN go mod download && go build
```

### 2. Build Arguments

```dockerfile
# Always use ARGs for versions
ARG GO_VERSION=1.21.0
ARG TARGETARCH

# Use in RUN commands
RUN curl -fsSL https://go.dev/dl/go${GO_VERSION}.linux-${TARGETARCH}.tar.gz
```

### 3. Security

```dockerfile
# Use specific tags, not :latest
FROM ubuntu:24.04

# Run as non-root when possible
USER nonroot:nonroot

# Use distroless for runtime
FROM gcr.io/distroless/static:nonroot
```

## Troubleshooting

### Build Cache Issues

```bash
# Clear build cache
docker builder prune -af

# Force rebuild without cache
make DOCKER_OPTS="--no-cache" docker/build/buildbase
```

### Multi-arch Build Failures

```bash
# Verify QEMU is installed
docker run --rm --privileged multiarch/qemu-user-static --reset -p yes

# Check buildx builder
docker buildx inspect --bootstrap
```

### Permission Errors in CI

```yaml
# Ensure correct permissions in workflow
permissions:
  contents: read
  packages: write
```

## References

- [BuildKit Documentation](https://github.com/moby/buildkit)
- [Docker Multi-platform Builds](https://docs.docker.com/build/building/multi-platform/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
