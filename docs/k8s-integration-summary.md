# Docker and Kubernetes Integration - Summary

This document summarizes the Docker and Kubernetes integration that has been added to the monorepo-go-example project.

## What's Been Added

### 1. Docker Integration

#### Dockerfiles (Already Created)
- ✅ `deployments/docker/user-service/Dockerfile` - Multi-stage build for user service
- ✅ `deployments/docker/order-service/Dockerfile` - Multi-stage build for order service  
- ✅ `deployments/docker/gateway/Dockerfile` - Multi-stage build for gateway

**Features:**
- Multi-stage builds (Go builder + Alpine runtime)
- Non-root user (uid: 1000)
- Security hardening (read-only filesystem, no privilege escalation)
- Health checks
- Minimal image size

#### Docker Compose (Already Created)
- ✅ `deployments/docker-compose.yml` - Complete local development stack with:
  - PostgreSQL
  - Redis
  - User Service
  - Order Service
  - Gateway
  - Prometheus
  - Grafana
  - Jaeger

### 2. Kubernetes Manifests

#### Core Infrastructure
- ✅ `deployments/k8s/namespace.yaml` - Namespace configuration
- ✅ `deployments/k8s/configmap.yaml` - ConfigMaps and Secrets for configuration
- ✅ `deployments/k8s/rbac.yaml` - ServiceAccount, Role, and RoleBinding
- ✅ `deployments/k8s/postgres.yaml` - PostgreSQL StatefulSet with persistent storage

#### Microservices
- ✅ `deployments/k8s/user-service.yaml` - User Service Deployment, Service, and HPA
- ✅ `deployments/k8s/order-service.yaml` - Order Service Deployment, Service, and HPA
- ✅ `deployments/k8s/gateway.yaml` - Gateway Deployment, Service (LoadBalancer), and HPA

#### Networking
- ✅ `deployments/k8s/ingress.yaml` - Ingress configuration with TLS support

**Features:**
- Horizontal Pod Autoscaling (HPA) for all services
- Resource requests and limits
- Liveness and readiness probes
- Security contexts (non-root, read-only filesystem)
- Prometheus annotations for monitoring
- Environment-based configuration via ConfigMaps/Secrets

### 3. Helm Charts

#### Chart Structure
```
charts/monorepo-go-example/
├── Chart.yaml                      # Chart metadata
├── values.yaml                     # Default configuration values
├── README.md                       # Chart documentation
├── .helmignore                     # Ignore patterns
└── templates/
    ├── _helpers.tpl                # Template helper functions
    ├── NOTES.txt                   # Post-installation notes
    ├── namespace.yaml              # Namespace template
    ├── serviceaccount.yaml         # ServiceAccount template
    ├── rbac.yaml                   # RBAC templates
    ├── configmap.yaml              # ConfigMap and Secret templates
    ├── postgres.yaml               # PostgreSQL StatefulSet template
    ├── user-service.yaml           # User Service templates
    ├── order-service.yaml          # Order Service templates
    ├── gateway.yaml                # Gateway templates
    └── ingress.yaml                # Ingress template
```

#### Key Features
- **Parameterized Configuration**: All values customizable via `values.yaml`
- **Conditional Resources**: Enable/disable components (ingress, postgresql, services)
- **Template Helpers**: Reusable template functions
- **Security Defaults**: Secure defaults with security contexts
- **Autoscaling**: HPA configuration for all services
- **Monitoring**: Prometheus integration
- **Documentation**: Comprehensive README with examples

### 4. Documentation

- ✅ `docs/docker-kubernetes.md` - Comprehensive guide covering:
  - Docker integration and best practices
  - Kubernetes deployment instructions
  - Helm chart usage
  - Local development setup (Docker Compose, Minikube, kind)
  - Production deployment guide
  - Monitoring and observability
  - Troubleshooting tips

- ✅ `charts/monorepo-go-example/README.md` - Helm chart documentation with:
  - Installation instructions
  - Configuration parameters
  - Usage examples
  - Troubleshooting

## Quick Start

### Local Development with Docker Compose

```bash
# Start all services
docker-compose -f deployments/docker-compose.yml up -d

# Access services
curl http://localhost:8080/v1/users
```

### Kubernetes Deployment

#### Option 1: Using kubectl

```bash
# Deploy all manifests
kubectl apply -f deployments/k8s/namespace.yaml
kubectl apply -f deployments/k8s/configmap.yaml
kubectl apply -f deployments/k8s/rbac.yaml
kubectl apply -f deployments/k8s/postgres.yaml
kubectl apply -f deployments/k8s/user-service.yaml
kubectl apply -f deployments/k8s/order-service.yaml
kubectl apply -f deployments/k8s/gateway.yaml

# Check status
kubectl get pods -n monorepo
```

#### Option 2: Using Helm (Recommended)

```bash
# Install with default values
helm install monorepo ./charts/monorepo-go-example

# Install with custom values
helm install monorepo ./charts/monorepo-go-example \
  --set secrets.database.password=mySecurePassword \
  --set gateway.service.type=ClusterIP

# Upgrade
helm upgrade monorepo ./charts/monorepo-go-example

# Uninstall
helm uninstall monorepo -n monorepo
```

### Using Makefile

```bash
# Build Docker images
make docker-build

# Push Docker images
make docker-push

# Deploy to Kubernetes
make k8s-deploy

# Install Helm chart
make helm-install

# Upgrade Helm release
make helm-upgrade

# Uninstall Helm release
make helm-uninstall
```

## Architecture Overview

### Container Architecture

```
┌─────────────────────────────────────────────────────┐
│                   Container Images                  │
├─────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌───────────┐ │
│  │ User Service │  │Order Service │  │  Gateway  │ │
│  │   (Alpine)   │  │   (Alpine)   │  │ (Alpine)  │ │
│  │  Multi-stage │  │  Multi-stage │  │Multi-stage│ │
│  │   Non-root   │  │   Non-root   │  │ Non-root  │ │
│  └──────────────┘  └──────────────┘  └───────────┘ │
└─────────────────────────────────────────────────────┘
```

### Kubernetes Architecture

```
┌────────────────────────────────────────────────────┐
│                  Ingress (Optional)                │
│             api.monorepo.example.com               │
│                   (NGINX + TLS)                    │
└────────────────────┬───────────────────────────────┘
                     │
┌────────────────────▼───────────────────────────────┐
│          Gateway (LoadBalancer/ClusterIP)          │
│                 Replicas: 2-15 (HPA)               │
│          Ports: 80 (HTTP), 9090 (gRPC)            │
└──────┬─────────────────────────────────┬───────────┘
       │                                 │
┌──────▼──────────┐            ┌─────────▼──────────┐
│  User Service   │            │  Order Service     │
│ Replicas: 2-10  │            │  Replicas: 2-10    │
│   (ClusterIP)   │            │   (ClusterIP)      │
│ Ports: 8081/9091│◄───────────┤ Ports: 8082/9092  │
└──────┬──────────┘            └─────────┬──────────┘
       │                                 │
       └────────────┬────────────────────┘
                    │
         ┌──────────▼──────────┐
         │   PostgreSQL        │
         │   (StatefulSet)     │
         │  Persistent Volume  │
         │    Port: 5432       │
         └─────────────────────┘
```

### Component Details

| Component | Type | Replicas | Ports | Autoscaling |
|-----------|------|----------|-------|-------------|
| Gateway | Deployment | 3 (2-15) | 80, 9090 | ✅ HPA |
| User Service | Deployment | 3 (2-10) | 8081, 9091 | ✅ HPA |
| Order Service | Deployment | 3 (2-10) | 8082, 9092 | ✅ HPA |
| PostgreSQL | StatefulSet | 1 | 5432 | ❌ |

## Configuration Management

### Environment Variables

All services are configured via environment variables sourced from:

1. **ConfigMaps** (`monorepo-config`):
   - `database.host`
   - `database.port`
   - `database.name`
   - `database.ssl_mode`
   - `log.level`
   - `log.format`
   - `server.mode`

2. **Secrets** (`monorepo-secrets`):
   - `database.user`
   - `database.password`

### Helm Values

Key configurable values in `values.yaml`:

```yaml
# Global settings
global:
  imageRegistry: ghcr.io

# PostgreSQL
postgresql:
  enabled: true
  persistence:
    size: 10Gi

# Services
userService:
  enabled: true
  replicaCount: 3
  autoscaling:
    enabled: true
    minReplicas: 2
    maxReplicas: 10

# Ingress
ingress:
  enabled: false
  hosts:
    - host: api.monorepo.example.com
```

## Resource Management

### Resource Requests and Limits

All services have defined resource constraints:

```yaml
resources:
  requests:
    memory: "128Mi"
    cpu: "100m"
  limits:
    memory: "256Mi"
    cpu: "200m"
```

### Horizontal Pod Autoscaling

HPA configuration based on CPU and memory:

```yaml
metrics:
- type: Resource
  resource:
    name: cpu
    target:
      type: Utilization
      averageUtilization: 70
- type: Resource
  resource:
    name: memory
    target:
      type: Utilization
      averageUtilization: 80
```

## Security Features

### Container Security

- ✅ **Non-root user**: All containers run as uid 1000
- ✅ **Read-only filesystem**: Root filesystem is read-only
- ✅ **No privilege escalation**: `allowPrivilegeEscalation: false`
- ✅ **Dropped capabilities**: All Linux capabilities dropped
- ✅ **Minimal base images**: Alpine Linux for small attack surface

### Kubernetes Security

- ✅ **RBAC**: ServiceAccount with minimal permissions
- ✅ **Namespace isolation**: Dedicated namespace
- ✅ **Secret management**: Sensitive data in Secrets
- ✅ **Network policies**: Ready for network policy implementation
- ✅ **Security contexts**: Pod and container security contexts

### TLS/SSL

- ✅ **Ingress TLS**: Optional TLS termination at Ingress
- ✅ **cert-manager integration**: Automatic certificate management
- ✅ **gRPC encryption**: Ready for mTLS between services

## Monitoring and Observability

### Prometheus Integration

All services expose metrics:

```yaml
annotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "8081"
  prometheus.io/path: "/metrics"
```

### Health Checks

Each service has:
- **Liveness probe**: `/health` endpoint
- **Readiness probe**: `/ready` endpoint

### Logging

- Structured JSON logs to stdout
- Aggregation via Kubernetes logging stack
- Jaeger tracing integration (in docker-compose)

## CI/CD Integration

The existing GitHub Actions workflow (`.github/workflows/ci.yml`) includes:

1. **Build Stage**: Docker image building
2. **Push Stage**: Push to container registry
3. **Deploy Stage**: Helm deployment to Kubernetes

## Production Readiness

### Checklist

- ✅ Multi-stage Docker builds
- ✅ Security hardening
- ✅ Health checks
- ✅ Resource limits
- ✅ Horizontal autoscaling
- ✅ Persistent storage for database
- ✅ Configuration management
- ✅ Secret management
- ✅ RBAC
- ✅ Monitoring integration
- ✅ Ingress with TLS
- ✅ Documentation

### Production Recommendations

1. **Change image tags from `latest`** to specific versions
2. **Set secure database passwords** in production secrets
3. **Configure storage class** for persistent volumes
4. **Enable Ingress with valid SSL certificates**
5. **Set up monitoring and alerting**
6. **Configure database backups**
7. **Implement network policies**
8. **Set up log aggregation**
9. **Regular security scanning**
10. **Disaster recovery plan**

## Next Steps

### Recommended Additions

1. **Service Mesh**: Consider Istio or Linkerd for advanced traffic management
2. **Network Policies**: Implement network segmentation
3. **Pod Disruption Budgets**: Ensure high availability during updates
4. **Backup/Restore**: Automated PostgreSQL backups
5. **Secrets Management**: Integration with external secret stores (Vault, AWS Secrets Manager)
6. **GitOps**: ArgoCD or Flux for declarative deployments
7. **Policy Enforcement**: OPA/Gatekeeper for policy as code
8. **Cost Optimization**: Resource right-sizing and cluster autoscaling

## Support and Resources

- **Documentation**: See `docs/docker-kubernetes.md` for detailed guide
- **Helm Chart**: See `charts/monorepo-go-example/README.md`
- **Main README**: See `README.md` for project overview
- **Makefile**: Run `make help` for available commands

## Testing

### Local Testing

```bash
# With Docker Compose
docker-compose -f deployments/docker-compose.yml up -d
curl http://localhost:8080/v1/users

# With Minikube
minikube start
make docker-build
helm install monorepo ./charts/monorepo-go-example
kubectl port-forward svc/gateway 8080:80 -n monorepo
curl http://localhost:8080/v1/users
```

### Kubernetes Validation

```bash
# Check all resources
kubectl get all -n monorepo

# Check HPA
kubectl get hpa -n monorepo

# Check logs
kubectl logs -l app=user-service -n monorepo

# Test connectivity
kubectl run -it --rm debug --image=alpine --restart=Never -- sh
apk add curl
curl http://user-service.monorepo.svc.cluster.local:8081/health
```

## Summary

The monorepo-go-example project now has complete Docker and Kubernetes integration with:

- **Production-ready Docker images** with security best practices
- **Complete Kubernetes manifests** for all services
- **Comprehensive Helm charts** for easy deployment
- **Horizontal autoscaling** for dynamic workload management
- **Security hardening** at container and Kubernetes levels
- **Monitoring integration** with Prometheus
- **Detailed documentation** for both development and production

You can now deploy this application to any Kubernetes cluster using either kubectl directly or Helm charts, with full support for local development using Docker Compose or Minikube.

## License

Copyright (C) 2025 Kevin Diu <kevindiujp@gmail.com>

Licensed under the Apache License, Version 2.0.