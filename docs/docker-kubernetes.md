# Docker and Kubernetes Integration Guide

This guide provides comprehensive information about Docker and Kubernetes integration for the monorepo-go-example project.

## Table of Contents

1. [Docker Integration](#docker-integration)
2. [Kubernetes Integration](#kubernetes-integration)
3. [Helm Charts](#helm-charts)
4. [Local Development](#local-development)
5. [Production Deployment](#production-deployment)
6. [Monitoring and Observability](#monitoring-and-observability)

## Docker Integration

### Docker Images

The project includes Dockerfiles for all microservices:

- **User Service**: `deployments/docker/user-service/Dockerfile`
- **Order Service**: `deployments/docker/order-service/Dockerfile`
- **Gateway**: `deployments/docker/gateway/Dockerfile`

### Multi-Stage Build

All Dockerfiles use multi-stage builds for optimal image size:

```dockerfile
# Stage 1: Build
FROM golang:1.21-alpine AS builder
...

# Stage 2: Runtime
FROM alpine:latest
...
```

### Building Docker Images

```bash
# Build all images
make docker-build

# Build specific service
docker build -t ghcr.io/kevindiu/monorepo-go-example/user-service:latest \
  -f deployments/docker/user-service/Dockerfile .

# Build with version tag
docker build -t ghcr.io/kevindiu/monorepo-go-example/user-service:v1.0.0 \
  -f deployments/docker/user-service/Dockerfile .
```

### Pushing Docker Images

```bash
# Push all images
make docker-push

# Push specific image
docker push ghcr.io/kevindiu/monorepo-go-example/user-service:latest
```

### Docker Compose

For local development, use Docker Compose:

```bash
# Start all services
docker-compose -f deployments/docker-compose.yml up -d

# View logs
docker-compose -f deployments/docker-compose.yml logs -f

# Stop services
docker-compose -f deployments/docker-compose.yml down
```

### Image Security

All images follow security best practices:

- ✅ Non-root user (uid: 1000)
- ✅ Read-only root filesystem
- ✅ Minimal base image (Alpine)
- ✅ No privileged escalation
- ✅ Dropped all capabilities
- ✅ Health checks included

## Kubernetes Integration

### Architecture

The Kubernetes deployment consists of:

```
┌─────────────────────────────────────────┐
│          Ingress (Optional)             │
│       api.monorepo.example.com          │
└────────────────┬────────────────────────┘
                 │
┌────────────────▼────────────────────────┐
│      Gateway Service (LoadBalancer)     │
│         Replicas: 2-15 (HPA)            │
└──────┬─────────────────────┬────────────┘
       │                     │
┌──────▼────────┐   ┌────────▼───────────┐
│ User Service  │   │  Order Service     │
│ Replicas: 2-10│   │  Replicas: 2-10    │
└──────┬────────┘   └────────┬───────────┘
       │                     │
       └──────────┬──────────┘
                  │
         ┌────────▼────────┐
         │   PostgreSQL    │
         │  StatefulSet    │
         │  (Persistent)   │
         └─────────────────┘
```

### Kubernetes Manifests

Manifests are located in `deployments/k8s/`:

- `namespace.yaml`: Namespace configuration
- `configmap.yaml`: ConfigMaps and Secrets
- `rbac.yaml`: Service Account, Role, and RoleBinding
- `postgres.yaml`: PostgreSQL StatefulSet
- `user-service.yaml`: User Service Deployment and Service
- `order-service.yaml`: Order Service Deployment and Service
- `gateway.yaml`: Gateway Deployment and Service
- `ingress.yaml`: Ingress configuration (optional)

### Deploying to Kubernetes

#### Using kubectl

```bash
# Deploy all resources
make k8s-deploy

# Or manually
kubectl apply -f deployments/k8s/namespace.yaml
kubectl apply -f deployments/k8s/configmap.yaml
kubectl apply -f deployments/k8s/rbac.yaml
kubectl apply -f deployments/k8s/postgres.yaml
kubectl apply -f deployments/k8s/user-service.yaml
kubectl apply -f deployments/k8s/order-service.yaml
kubectl apply -f deployments/k8s/gateway.yaml
```

#### Update configuration

```bash
# Edit ConfigMap
kubectl edit configmap monorepo-config -n monorepo

# Edit Secret
kubectl edit secret monorepo-secrets -n monorepo

# Restart pods to pick up changes
kubectl rollout restart deployment/user-service -n monorepo
kubectl rollout restart deployment/order-service -n monorepo
kubectl rollout restart deployment/gateway -n monorepo
```

### Horizontal Pod Autoscaling

All services include HPA configuration:

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: user-service-hpa
spec:
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

View HPA status:

```bash
kubectl get hpa -n monorepo
```

### Resource Management

Each service has resource requests and limits:

```yaml
resources:
  requests:
    memory: "128Mi"
    cpu: "100m"
  limits:
    memory: "256Mi"
    cpu: "200m"
```

### Health Checks

Liveness and readiness probes are configured:

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8081
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /ready
    port: 8081
  initialDelaySeconds: 5
  periodSeconds: 5
```

## Helm Charts

### Chart Structure

```
charts/monorepo-go-example/
├── Chart.yaml           # Chart metadata
├── values.yaml          # Default values
├── README.md            # Chart documentation
├── .helmignore          # Ignore patterns
└── templates/
    ├── _helpers.tpl     # Template helpers
    ├── NOTES.txt        # Post-install notes
    ├── namespace.yaml
    ├── serviceaccount.yaml
    ├── rbac.yaml
    ├── configmap.yaml
    ├── postgres.yaml
    ├── user-service.yaml
    ├── order-service.yaml
    ├── gateway.yaml
    └── ingress.yaml
```

### Installing with Helm

```bash
# Install with default values
helm install monorepo ./charts/monorepo-go-example

# Install with custom values
helm install monorepo ./charts/monorepo-go-example \
  --set secrets.database.password=myPassword \
  --set gateway.service.type=ClusterIP

# Install from values file
helm install monorepo ./charts/monorepo-go-example -f custom-values.yaml
```

### Upgrading

```bash
# Upgrade release
helm upgrade monorepo ./charts/monorepo-go-example

# Upgrade with new values
helm upgrade monorepo ./charts/monorepo-go-example \
  --set userService.replicaCount=5
```

### Helm Values

Key configurable values:

```yaml
# PostgreSQL
postgresql:
  enabled: true
  persistence:
    size: 10Gi

# User Service
userService:
  replicaCount: 3
  autoscaling:
    enabled: true
    minReplicas: 2
    maxReplicas: 10

# Gateway
gateway:
  service:
    type: LoadBalancer  # or ClusterIP
  
# Ingress
ingress:
  enabled: false
  hosts:
    - host: api.example.com
```

## Local Development

### Development with Docker Compose

```bash
# Start all services
docker-compose -f deployments/docker-compose.yml up -d

# Access services
curl http://localhost:8080/v1/users  # Gateway
curl http://localhost:8081/v1/users  # User Service
curl http://localhost:8082/v1/orders # Order Service

# View logs
docker-compose logs -f user-service

# Rebuild and restart
docker-compose up -d --build user-service
```

### Development with Minikube

```bash
# Start minikube
minikube start --cpus=4 --memory=8192

# Enable addons
minikube addons enable ingress
minikube addons enable metrics-server

# Build images in minikube
eval $(minikube docker-env)
make docker-build

# Deploy
helm install monorepo ./charts/monorepo-go-example

# Access gateway
minikube service gateway -n monorepo

# Or port-forward
kubectl port-forward svc/gateway 8080:80 -n monorepo
```

### Development with kind

```bash
# Create cluster
kind create cluster --name monorepo

# Load images
kind load docker-image ghcr.io/kevindiu/monorepo-go-example/user-service:latest
kind load docker-image ghcr.io/kevindiu/monorepo-go-example/order-service:latest
kind load docker-image ghcr.io/kevindiu/monorepo-go-example/gateway:latest

# Deploy
helm install monorepo ./charts/monorepo-go-example \
  --set gateway.service.type=NodePort

# Access via port-forward
kubectl port-forward svc/gateway 8080:80 -n monorepo
```

## Production Deployment

### Prerequisites

1. Kubernetes cluster (GKE, EKS, AKS, etc.)
2. kubectl configured
3. Helm 3+ installed
4. Container registry access
5. SSL certificates (for Ingress)

### Production Checklist

- [ ] Update image tags from `latest` to specific versions
- [ ] Set secure database passwords in secrets
- [ ] Configure persistent storage class
- [ ] Enable Ingress with SSL/TLS
- [ ] Configure resource requests/limits
- [ ] Set up monitoring and alerting
- [ ] Configure backup for PostgreSQL
- [ ] Review security policies
- [ ] Configure network policies
- [ ] Set up logging aggregation

### Production Values

Create `production-values.yaml`:

```yaml
secrets:
  database:
    password: <SECURE_PASSWORD>

postgresql:
  persistence:
    storageClass: fast-ssd
    size: 100Gi

userService:
  image:
    tag: v1.0.0
  replicaCount: 5
  resources:
    requests:
      memory: "256Mi"
      cpu: "200m"
    limits:
      memory: "512Mi"
      cpu: "500m"

orderService:
  image:
    tag: v1.0.0
  replicaCount: 5
  resources:
    requests:
      memory: "256Mi"
      cpu: "200m"
    limits:
      memory: "512Mi"
      cpu: "500m"

gateway:
  image:
    tag: v1.0.0
  service:
    type: ClusterIP
  replicaCount: 5

ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
  hosts:
    - host: api.production.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: api-tls
      hosts:
        - api.production.example.com
```

Deploy:

```bash
helm install monorepo ./charts/monorepo-go-example \
  -f production-values.yaml \
  --namespace production \
  --create-namespace
```

### Rolling Updates

```bash
# Update image version
helm upgrade monorepo ./charts/monorepo-go-example \
  --set userService.image.tag=v1.1.0 \
  --reuse-values

# Check rollout status
kubectl rollout status deployment/user-service -n monorepo

# Rollback if needed
helm rollback monorepo 1
```

### Blue-Green Deployment

```bash
# Deploy new version (green)
helm install monorepo-green ./charts/monorepo-go-example \
  --set userService.image.tag=v2.0.0 \
  --set namespace.name=monorepo-green

# Test green environment
kubectl port-forward svc/gateway 8081:80 -n monorepo-green

# Switch traffic (update ingress)
helm upgrade monorepo ./charts/monorepo-go-example \
  --set namespace.name=monorepo-green

# Remove old blue environment
helm uninstall monorepo-blue
```

## Monitoring and Observability

### Prometheus Metrics

All services expose Prometheus metrics at `/metrics`:

```yaml
annotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "8081"
  prometheus.io/path: "/metrics"
```

### Logging

Structured JSON logging to stdout:

```bash
# View logs
kubectl logs -l app=user-service -n monorepo

# Follow logs
kubectl logs -f deployment/user-service -n monorepo

# View logs from all pods
kubectl logs -l app=user-service -n monorepo --all-containers=true
```

### Tracing

Jaeger integration (in docker-compose):

```bash
# Access Jaeger UI
http://localhost:16686
```

### Health Checks

```bash
# Check service health
kubectl exec -it <pod-name> -n monorepo -- wget -qO- localhost:8081/health

# Check readiness
kubectl exec -it <pod-name> -n monorepo -- wget -qO- localhost:8081/ready
```

### Debugging

```bash
# Get pod details
kubectl describe pod <pod-name> -n monorepo

# Execute commands in pod
kubectl exec -it <pod-name> -n monorepo -- sh

# Port forward for debugging
kubectl port-forward <pod-name> 8081:8081 -n monorepo

# View events
kubectl get events -n monorepo --sort-by='.lastTimestamp'
```

## CI/CD Integration

GitHub Actions workflow includes Docker and Kubernetes deployment:

```yaml
- name: Build Docker Images
  run: make docker-build

- name: Push Docker Images
  run: make docker-push

- name: Deploy to Kubernetes
  run: |
    helm upgrade --install monorepo ./charts/monorepo-go-example \
      --set userService.image.tag=${{ github.sha }} \
      --wait
```

## Best Practices

### Docker

1. Use multi-stage builds
2. Run as non-root user
3. Use specific base image tags
4. Minimize layers
5. Use .dockerignore
6. Scan images for vulnerabilities
7. Use health checks

### Kubernetes

1. Use namespaces for isolation
2. Set resource requests/limits
3. Use liveness and readiness probes
4. Enable horizontal pod autoscaling
5. Use ConfigMaps and Secrets
6. Implement RBAC
7. Use network policies
8. Enable monitoring and logging
9. Regular backups
10. Use Helm for deployment management

### Security

1. Use least privilege RBAC
2. Enable pod security policies
3. Scan images for vulnerabilities
4. Use secrets for sensitive data
5. Enable network policies
6. Use TLS/SSL everywhere
7. Regular security updates
8. Audit logging

## Troubleshooting

### Common Issues

#### Pods not starting

```bash
kubectl get pods -n monorepo
kubectl describe pod <pod-name> -n monorepo
kubectl logs <pod-name> -n monorepo
```

#### Database connection issues

```bash
# Check if postgres is running
kubectl get pods -l app=postgres -n monorepo

# Test connection
kubectl exec -it postgres-0 -n monorepo -- psql -U postgres -d monorepo -c '\l'

# Check config
kubectl get configmap monorepo-config -n monorepo -o yaml
kubectl get secret monorepo-secrets -n monorepo -o yaml
```

#### Service not accessible

```bash
# Check service
kubectl get svc -n monorepo
kubectl describe svc gateway -n monorepo

# Check endpoints
kubectl get endpoints -n monorepo

# Test connectivity
kubectl run -it --rm debug --image=alpine --restart=Never -- sh
apk add curl
curl http://user-service.monorepo.svc.cluster.local:8081/health
```

#### HPA not scaling

```bash
# Check HPA status
kubectl get hpa -n monorepo
kubectl describe hpa user-service-hpa -n monorepo

# Check metrics server
kubectl top nodes
kubectl top pods -n monorepo
```

## Additional Resources

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Helm Documentation](https://helm.sh/docs/)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Project README](../../README.md)

## License

Copyright (C) 2025 Kevin Diu <kevindiujp@gmail.com>

Licensed under the Apache License, Version 2.0.