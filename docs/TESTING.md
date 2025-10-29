# Testing Guide

This document describes the testing strategy for the monorepo-go-example project.

## Test Organization

The project uses three types of tests:

### 1. Unit Tests
Located alongside the code they test (`*_test.go` files).

**Coverage:**
- `internal/log` - Logger configuration and field constructors
- `internal/errors` - Error codes and wrapping
- `internal/config` - Configuration loading and validation
- `pkg/order/service` - Order service business logic
- `pkg/user/service` - User service business logic

**Run unit tests:**
```bash
go test ./internal/...
go test ./pkg/order/service/...
go test ./pkg/user/service/...
```

### 2. Integration Tests
Tests that require external dependencies (database, message queue, etc.).

**Coverage:**
- `pkg/order/repository` - Database operations for orders
- `pkg/user/repository` - Database operations for users

**Run integration tests:**
```bash
# Start test database first
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
go test -v -tags=integration ./pkg/order/repository/...
go test -v -tags=integration ./pkg/user/repository/...

# Cleanup
docker-compose -f docker-compose.test.yml down
```

### 3. End-to-End Tests
Full system tests with all services running.

**Coverage:**
- `tests/e2e` - Complete workflows across services

**Run E2E tests:**
```bash
# Start all services
docker-compose up -d

# Wait for services to be ready
sleep 10

# Run E2E tests
go test -v ./tests/e2e/...

# Cleanup
docker-compose down
```

## Test Commands

### Run All Tests
```bash
make test
```

### Run Only Unit Tests
```bash
make test-unit
```

### Run Integration Tests
```bash
make test-integration
```

### Run E2E Tests
```bash
make test-e2e
```

### Run Tests with Coverage
```bash
make test-coverage
```

### Run Tests in Short Mode (Skip Long Tests)
```bash
go test -short ./...
```

## Writing Tests

### Unit Test Example
```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   "test",
            want:    "TEST",
            wantErr: false,
        },
        {
            name:    "empty input",
            input:   "",
            want:    "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MyFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("MyFunction() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("MyFunction() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Integration Test Example
```go
func TestRepository_Create(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    db := setupTestDB(t)
    defer db.Close()

    repo := NewRepository(db)
    
    // Test logic here
}
```

### E2E Test Example
```go
func TestOrderWorkflow(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test in short mode")
    }

    cluster := NewTestCluster(t)
    defer cluster.Cleanup()

    ctx := context.Background()
    
    // Wait for services
    if err := cluster.WaitForHealthy(ctx, 10*time.Second); err != nil {
        t.Fatalf("Services not ready: %v", err)
    }

    // Test logic here
}
```

## Mock Usage

Services use mock implementations for testing without external dependencies:

```go
type mockRepository struct {
    data map[string]*Entity
}

func (m *mockRepository) Create(ctx context.Context, entity *Entity) error {
    m.data[entity.ID] = entity
    return nil
}

func TestService(t *testing.T) {
    repo := &mockRepository{data: make(map[string]*Entity)}
    svc := NewService(repo)
    
    // Test service with mock repository
}
```

## CI/CD Integration

### GitHub Actions Example
```yaml
name: Tests

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '1.21'
      - run: make test-unit

  integration-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '1.21'
      - run: make test-integration

  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '1.21'
      - run: docker-compose up -d
      - run: make test-e2e
      - run: docker-compose down
```

## Test Coverage Goals

- **Unit Tests:** > 80% coverage for business logic
- **Integration Tests:** All repository methods tested
- **E2E Tests:** Critical user workflows covered

## Best Practices

1. **Use table-driven tests** for multiple test cases
2. **Mock external dependencies** in unit tests
3. **Use `t.Skip()`** for tests requiring external services in CI
4. **Clean up resources** using `defer` or `t.Cleanup()`
5. **Use meaningful test names** describing what is being tested
6. **Test both success and error cases**
7. **Keep tests fast** - unit tests should run in milliseconds
8. **Isolate tests** - tests should not depend on each other
9. **Use `testing.Short()`** to skip long-running tests
10. **Document test requirements** (database, services, etc.)

## Debugging Tests

### Run a specific test
```bash
go test -v -run TestMyFunction ./pkg/...
```

### Run tests with race detection
```bash
go test -race ./...
```

### Run tests with verbose output
```bash
go test -v ./...
```

### Generate coverage report
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Troubleshooting

### Tests fail with "connection refused"
- Ensure required services are running (database, etc.)
- Check service health with `docker-compose ps`
- Wait longer for services to start

### Tests timeout
- Increase test timeout: `go test -timeout 30s`
- Check for deadlocks or infinite loops
- Verify external services are responsive

### Flaky tests
- Add retry logic for network operations
- Use proper synchronization (channels, mutexes)
- Avoid time-based assertions
