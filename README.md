# Monorepo Go Example

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![CI Pipeline](https://github.com/kevindiu/monorepo-go-example/workflows/CI%20Pipeline/badge.svg)](https://github.com/kevindiu/monorepo-go-example/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/kevindiu/monorepo-go-example)](https://goreportcard.com/report/github.com/kevindiu/monorepo-go-example)
[![GoDoc](https://godoc.org/github.com/kevindiu/monorepo-go-example?status.svg)](https://godoc.org/github.com/kevindiu/monorepo-go-example)

A comprehensive Go monorepo example demonstrating microservices architecture with gRPC, REST APIs, database integration, Kubernetes deployment, and comprehensive testing infrastructure.

## ✨ Highlights

This project features a **production-grade CI/CD system** with:

- 🐳 **Multi-stage Docker builds** - Buildbase → CI Container → Service Images
- 📦 **Version management** - Centralized tool versions in `versions/` directory
- ⚡ **BuildKit optimizations** - Multi-level caching (apt, go-pkg, go-build)
- 🌍 **Multi-architecture** - Support for linux/amd64 and linux/arm64
- 🔄 **Reusable workflows** - DRY principle for GitHub Actions
- 🛠️ **Pre-built CI container** - All tools pre-installed for fast CI runs
- 📊 **Advanced caching** - GitHub Actions cache + Registry cache + BuildKit

**Read more**: [CI/CD Implementation](docs/CICD_IMPLEMENTATION.md) | [Docker Build System](docs/DOCKER_BUILD_SYSTEM.md)

## 🏗️ Architecture

This monorepo contains multiple microservices built with Go, demonstrating:

- **Microservices Architecture**: Independent, scalable services
- **gRPC & REST APIs**: High-performance communication with HTTP gateway
- **Database Integration**: PostgreSQL with migration support
- **Kubernetes Ready**: Complete K8s manifests and Helm charts
- **Comprehensive Testing**: Unit, integration, and E2E tests
- **CI/CD Pipeline**: GitHub Actions workflow
- **Development Tools**: Makefile, Docker, proto generation

### Services

1. **User Service** (`cmd/user-service/`) - User management operations
2. **Order Service** (`cmd/order-service/`) - Order processing operations  
3. **Gateway Service** (`cmd/gateway/`) - API gateway and routing

## 📁 Project Structure

```
├── cmd/                    # Service entry points
│   ├── user-service/       # User service main
│   ├── order-service/      # Order service main
│   └── gateway/            # Gateway service main
├── pkg/                    # Public packages (business logic)
│   ├── user/               # User domain logic
│   ├── order/              # Order domain logic
│   └── gateway/            # Gateway logic
├── internal/               # Private packages (shared utilities)
│   ├── config/             # Configuration management
│   ├── db/                 # Database utilities
│   ├── errors/             # Error handling
│   ├── log/                # Logging utilities
│   └── middleware/         # gRPC/HTTP middleware
├── apis/                   # API definitions and generated code
│   ├── proto/              # Protocol buffer definitions
│   └── grpc/               # Generated gRPC code
├── deployments/            # Deployment configurations
│   ├── docker/             # Dockerfiles
│   ├── k8s/                # Kubernetes manifests
│   └── docker-compose.yml  # Local development
├── charts/                 # Helm charts
├── tests/                  # Integration and E2E tests
├── docs/                   # Documentation
├── hack/                   # Build and development scripts
└── .github/                # GitHub Actions workflows
```

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+**
- **PostgreSQL 13+**
- **Docker** (for containerized development)
- **kubectl** (for Kubernetes deployment)
- **Helm 3+** (optional, for Helm charts)

### Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/kevindiu/monorepo-go-example.git
   cd monorepo-go-example
   ```

2. **Setup development environment**
   ```bash
   make dev-setup
   ```
   This will install dependencies, development tools, and generate protobuf code.

3. **Start PostgreSQL** (using Docker)
   ```bash
   docker run --name postgres-dev -e POSTGRES_DB=monorepo -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d postgres:13
   ```

4. **Build all services**
   ```bash
   make build
   ```

5. **Run services**
   ```bash
   # Terminal 1: User Service
   make run-user-service
   
   # Terminal 2: Order Service  
   make run-order-service
   
   # Terminal 3: Gateway
   make run-gateway
   ```

### Using Docker Compose

For local development with all dependencies:

```bash
make docker-run
```

This starts:
- PostgreSQL database
- All microservices
- Monitoring stack (optional)

## 🔧 Development

### Available Make Targets

```bash
make help                   # Show all available targets
make dev-setup             # Setup development environment
make build                 # Build all services
make test                  # Run all tests
make test-unit             # Run unit tests only
make test-integration      # Run integration tests
make lint                  # Run linter
make fmt                   # Format code
make proto                 # Generate protobuf code
make docker-build          # Build Docker images
make k8s-deploy            # Deploy to Kubernetes
make helm-install          # Install Helm chart
```

### Code Generation

Protocol buffer code is generated using [buf](https://buf.build/):

```bash
make proto
```

### Testing

```bash
# Run all tests
make test

# Run only unit tests
make test-unit

# Run only integration tests
make test-integration
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Run go vet
make vet
```

## 🐳 Docker

### Building Images

```bash
make docker-build
```

### Running with Docker Compose

```bash
# Start all services
make docker-run

# Stop all services
make docker-stop
```

## ☸️ Kubernetes Deployment

### Using kubectl

```bash
# Deploy all services
make k8s-deploy

# Remove deployment
make k8s-undeploy
```

### Using Helm

```bash
# Install chart
make helm-install

# Upgrade chart
make helm-upgrade

# Uninstall chart
make helm-uninstall
```

## 📚 API Documentation

### gRPC APIs

- **User Service**: Port 9091
  - `CreateUser`
  - `GetUser`
  - `ListUsers`
  - `UpdateUser`
  - `DeleteUser`

- **Order Service**: Port 9092
  - `CreateOrder`
  - `GetOrder`
  - `ListOrders`
  - `UpdateOrderStatus`
  - `CancelOrder`

### REST APIs (via Gateway)

All gRPC services are exposed via REST through the gateway on port 8080:

- `POST /v1/users` - Create user
- `GET /v1/users/{id}` - Get user
- `GET /v1/users` - List users
- `PUT /v1/users/{id}` - Update user
- `DELETE /v1/users/{id}` - Delete user

- `POST /v1/orders` - Create order
- `GET /v1/orders/{id}` - Get order
- `GET /v1/orders` - List orders
- `PUT /v1/orders/{id}/status` - Update order status
- `DELETE /v1/orders/{id}` - Cancel order

## 🧪 Testing

### Test Structure

- **Unit Tests**: `*_test.go` files alongside source code
- **Integration Tests**: `tests/integration/`
- **E2E Tests**: `tests/e2e/`

### Running Tests

```bash
# All tests
make test

# Unit tests only
make test-unit

# Integration tests only  
make test-integration
```

## 🏗️ Architecture Principles

This project follows several architectural principles:

1. **Clean Architecture**: Clear separation of concerns
2. **Domain-Driven Design**: Business logic in domain packages
3. **Dependency Injection**: Testable and maintainable code
4. **Error Handling**: Structured error handling with codes
5. **Observability**: Comprehensive logging and metrics
6. **Configuration**: Environment-based configuration
7. **Testing**: High test coverage with multiple test types

## 🧪 Testing

### Test Structure

The project includes comprehensive testing:

- **Unit Tests** (`*_test.go`) - Fast, isolated tests for individual components
- **Integration Tests** (`pkg/*/repository/*_test.go`) - Tests with database
- **E2E Tests** (`tests/e2e/`) - Full system tests with all services

### Running Tests

```bash
# Run all tests
make test

# Run only unit tests
make test-unit

# Run only integration tests (requires database)
make test-integration

# Run E2E tests (requires running services)
make test-e2e

# Generate coverage report
make test-coverage
```

See [Testing Guide](docs/TESTING.md) for detailed information.

## 🔄 CI/CD

This project features a **production-grade CI/CD system** with sophisticated build automation and caching strategies.

### Docker Build System

The build system uses a multi-stage hierarchy:

```
Service Images (user-service, config-service)
    ↓ FROM
CI Container (all build tools pre-installed)
    ↓ FROM  
Buildbase (Ubuntu + Go)
```

**Key Features**:
- 🐳 Multi-stage base images for optimal caching
- 📦 Version management in `versions/` directory
- ⚡ BuildKit cache mounts (apt, go-pkg, go-build)
- 🌍 Multi-architecture builds (amd64, arm64)
- 🔄 Reusable GitHub Actions workflows
- 📊 Advanced caching (GHA + Registry + BuildKit)

### Build Commands

```bash
# Build CI container (all tools pre-installed)
make docker/build/ci-container

# Build all images
make docker/build/all

# Build with remote caching
make REMOTE=true docker/build/ci-container

# Build for specific platform
make PLATFORM=linux/arm64 docker/build/buildbase
```

### GitHub Actions Workflows

- **`docker-buildbase.yml`** - Build foundation image (weekly)
- **`docker-ci-container.yml`** - Build CI container (daily)  
- **`ci.yml`** - Main CI pipeline using pre-built container
- **`_docker-build.yml`** - Reusable build workflow
- **`pr-validation.yml`** - Fast PR validation
- **`e2e-tests.yml`** - End-to-end testing
- **`release.yml`** - Release automation

### CI Container Usage

All CI jobs run in a pre-built container with all tools installed:

```yaml
jobs:
  test:
    container:
      image: ghcr.io/kevindiu/monorepo-go-example-ci-container:latest
    steps:
      - run: make test  # Tools already installed!
```

**Benefits**:
- ⚡ 5x faster CI runs (no tool installation)
- 🎯 Consistent environment (local + CI)
- 📦 All tools pre-installed (go, protoc, buf, kubectl, etc.)

### Documentation

- 📖 [CI/CD Implementation](docs/CICD_IMPLEMENTATION.md) - Complete implementation guide
- 🐳 [Docker Build System](docs/DOCKER_BUILD_SYSTEM.md) - Docker build architecture
- 🛠️ [CI Container README](dockers/ci/base/README.md) - CI container documentation
- 📋 [CI/CD Guide](docs/CICD.md) - Workflows and usage

See [CI/CD Guide](docs/CICD.md) for detailed information.

## 📈 Monitoring

The project includes monitoring setup with:

- **Prometheus**: Metrics collection
- **Grafana**: Metrics visualization
- **Jaeger**: Distributed tracing
- **Structured Logging**: JSON-formatted logs

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes
4. Run tests: `make test`
5. Run linter: `make lint`
6. Commit your changes: `git commit -m 'Add amazing feature'`
7. Push to the branch: `git push origin feature/amazing-feature`
8. Open a Pull Request

## 📋 TODO

- [ ] Add monitoring and observability stack
- [ ] Implement service mesh (Istio) integration
- [ ] Add more comprehensive examples
- [ ] Create performance benchmarks
- [ ] Add security scanning
- [ ] Implement event sourcing example

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with ❤️ for the Go community
- Thanks to all contributors

---

## Support

If you have questions or need help, please:

1. Check the [documentation](docs/)
2. Open an [issue](https://github.com/kevindiu/monorepo-go-example/issues)
3. Start a [discussion](https://github.com/kevindiu/monorepo-go-example/discussions)

**Happy coding! 🚀**