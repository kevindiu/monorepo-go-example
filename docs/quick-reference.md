# Quick Reference Guide

## Common Commands

### Development

```bash
# Install dependencies
make deps

# Generate protobuf code
make proto

# Build all services
make build

# Build specific service
make build-user-service
make build-order-service
make build-gateway

# Run services locally
make run-user-service
make run-order-service
make run-gateway

# Run tests
make test
make test-unit
make test-integration

# Code quality
make lint
make fmt
```

### Docker

```bash
# Build Docker images
make docker-build

# Build specific image
docker build -t user-service:latest -f deployments/docker/user-service/Dockerfile .

# Run with Docker Compose
docker-compose -f deployments/docker-compose.yml up -d

# View logs
docker-compose -f deployments/docker-compose.yml logs -f user-service

# Stop services
docker-compose -f deployments/docker-compose.yml down

# Push images
make docker-push
```

### Kubernetes

```bash
# Deploy all manifests
make k8s-deploy

# Or manually
kubectl apply -f deployments/k8s/namespace.yaml
kubectl apply -f deployments/k8s/configmap.yaml
kubectl apply -f deployments/k8s/rbac.yaml
kubectl apply -f deployments/k8s/postgres.yaml
kubectl apply -f deployments/k8s/user-service.yaml
kubectl apply -f deployments/k8s/order-service.yaml
kubectl apply -f deployments/k8s/gateway.yaml

# Delete all resources
kubectl delete -f deployments/k8s/
```

### Helm

```bash
# Install chart
make helm-install

# Or manually
helm install monorepo ./charts/monorepo-go-example

# Upgrade
make helm-upgrade

# Uninstall
make helm-uninstall

# Install with custom values
helm install monorepo ./charts/monorepo-go-example -f custom-values.yaml

# Dry run
helm install monorepo ./charts/monorepo-go-example --dry-run --debug
```

### Kubernetes Operations

```bash
# Check pods
kubectl get pods -n monorepo

# Check services
kubectl get svc -n monorepo

# Check deployments
kubectl get deployments -n monorepo

# Check HPA
kubectl get hpa -n monorepo

# View logs
kubectl logs -f deployment/user-service -n monorepo
kubectl logs -f deployment/order-service -n monorepo
kubectl logs -f deployment/gateway -n monorepo

# Describe resources
kubectl describe pod <pod-name> -n monorepo
kubectl describe svc gateway -n monorepo

# Port forward
kubectl port-forward svc/gateway 8080:80 -n monorepo

# Execute commands in pod
kubectl exec -it <pod-name> -n monorepo -- sh

# Scale deployment
kubectl scale deployment user-service --replicas=5 -n monorepo

# Restart deployment
kubectl rollout restart deployment/user-service -n monorepo

# Check rollout status
kubectl rollout status deployment/user-service -n monorepo
```

## API Examples

### User Service (via Gateway)

```bash
# Create user
curl -X POST http://localhost:8080/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "age": 30
  }'

# Get user
curl http://localhost:8080/v1/users/{user-id}

# List users
curl http://localhost:8080/v1/users?page_size=10

# Update user
curl -X PUT http://localhost:8080/v1/users/{user-id} \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Smith",
    "email": "john.smith@example.com",
    "age": 31
  }'

# Delete user
curl -X DELETE http://localhost:8080/v1/users/{user-id}
```

### Order Service (via Gateway)

```bash
# Create order
curl -X POST http://localhost:8080/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-123",
    "items": [
      {
        "product_id": "prod-1",
        "quantity": 2,
        "price": 29.99
      },
      {
        "product_id": "prod-2",
        "quantity": 1,
        "price": 49.99
      }
    ]
  }'

# Get order
curl http://localhost:8080/v1/orders/{order-id}

# List orders
curl http://localhost:8080/v1/orders?page_size=10

# List user orders
curl http://localhost:8080/v1/orders?user_id={user-id}

# Update order status
curl -X PUT http://localhost:8080/v1/orders/{order-id}/status \
  -H "Content-Type: application/json" \
  -d '{
    "status": "shipped"
  }'

# Cancel order
curl -X POST http://localhost:8080/v1/orders/{order-id}/cancel
```

### Health Checks

```bash
# Gateway health
curl http://localhost:8080/health
curl http://localhost:8080/ready

# User service health (direct)
curl http://localhost:8081/health
curl http://localhost:8081/ready

# Order service health (direct)
curl http://localhost:8082/health
curl http://localhost:8082/ready
```

## Environment Variables

### Common Variables

```bash
# Server configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_GRPC_PORT=9090

# Database configuration
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=monorepo
DATABASE_USER=postgres
DATABASE_PASSWORD=changeme
DATABASE_SSL_MODE=disable

# Logging
LOG_LEVEL=info
LOG_FORMAT=json

# Gateway-specific
USER_SERVICE_ENDPOINT=localhost:9091
ORDER_SERVICE_ENDPOINT=localhost:9092
```

## Configuration Files

### Config File (YAML)

```yaml
# config.yaml
server:
  host: 0.0.0.0
  port: 8080
  grpc_port: 9090

database:
  host: localhost
  port: 5432
  name: monorepo
  user: postgres
  password: changeme
  ssl_mode: disable

log:
  level: info
  format: json
```

### Docker Compose Override

```yaml
# docker-compose.override.yml
version: '3.8'

services:
  user-service:
    environment:
      - LOG_LEVEL=debug
    ports:
      - "8081:8081"
```

## Troubleshooting

### Check Service Status

```bash
# Docker Compose
docker-compose ps

# Kubernetes
kubectl get pods -n monorepo
kubectl get events -n monorepo --sort-by='.lastTimestamp'
```

### View Logs

```bash
# Docker Compose
docker-compose logs -f user-service

# Kubernetes
kubectl logs -f deployment/user-service -n monorepo --tail=100
```

### Database Connection

```bash
# Docker Compose
docker-compose exec postgres psql -U postgres -d monorepo

# Kubernetes
kubectl exec -it postgres-0 -n monorepo -- psql -U postgres -d monorepo
```

### Common Issues

1. **Port already in use**
   ```bash
   # Find process using port
   lsof -i :8080
   # Kill process
   kill -9 <PID>
   ```

2. **Database connection failed**
   - Check if PostgreSQL is running
   - Verify connection string
   - Check network connectivity

3. **gRPC connection failed**
   - Verify service is running
   - Check endpoint configuration
   - Check firewall/network policies

4. **Pod not starting in K8s**
   ```bash
   kubectl describe pod <pod-name> -n monorepo
   kubectl logs <pod-name> -n monorepo
   ```

## Monitoring

### Prometheus Metrics

```bash
# User service metrics
curl http://localhost:8081/metrics

# Order service metrics
curl http://localhost:8082/metrics

# Gateway metrics
curl http://localhost:8080/metrics
```

### Grafana (Docker Compose)

Access at: http://localhost:3000
- Username: admin
- Password: admin

### Jaeger (Docker Compose)

Access at: http://localhost:16686

## Development Tips

1. **Hot Reload**: Use `air` or `fresh` for auto-reload during development
2. **Database Migrations**: Place new migrations in `hack/db/migrations/`
3. **Proto Changes**: Run `make proto` after modifying .proto files
4. **Format Code**: Run `make fmt` before committing
5. **Lint**: Run `make lint` to catch issues early

## Production Checklist

- [ ] Update image tags from `latest` to specific versions
- [ ] Set secure database passwords
- [ ] Configure persistent storage class
- [ ] Enable TLS/SSL on Ingress
- [ ] Set appropriate resource limits
- [ ] Configure monitoring and alerting
- [ ] Set up log aggregation
- [ ] Configure backup for database
- [ ] Review and apply security policies
- [ ] Test disaster recovery procedures

## Useful Kubectl Plugins

```bash
# Install krew (kubectl plugin manager)
# https://krew.sigs.k8s.io/docs/user-guide/setup/install/

# Useful plugins
kubectl krew install ns      # Switch namespaces
kubectl krew install ctx     # Switch contexts
kubectl krew install tail    # Tail logs
kubectl krew install tree    # Resource hierarchy
```

## Resources

- [Kubernetes Docs](https://kubernetes.io/docs/)
- [Helm Docs](https://helm.sh/docs/)
- [gRPC Go](https://grpc.io/docs/languages/go/)
- [Protocol Buffers](https://developers.google.com/protocol-buffers)
- [Viper Config](https://github.com/spf13/viper)
- [Zap Logger](https://github.com/uber-go/zap)

---

For more details, see the [main README](../README.md) and [documentation](../docs/).
