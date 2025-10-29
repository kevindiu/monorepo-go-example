# Monorepo Go Example Helm Chart

This Helm chart deploys the monorepo-go-example microservices application on a Kubernetes cluster.

## Prerequisites

- Kubernetes 1.20+
- Helm 3.0+
- PV provisioner support in the underlying infrastructure (for PostgreSQL persistence)

## Installing the Chart

To install the chart with the release name `my-release`:

```bash
helm install my-release ./charts/monorepo-go-example
```

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```bash
helm uninstall my-release -n monorepo
```

## Configuration

The following table lists the configurable parameters of the chart and their default values.

### Global Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `global.imageRegistry` | Global Docker image registry | `ghcr.io` |
| `global.imagePullSecrets` | Global Docker registry secret names as an array | `[]` |
| `global.storageClass` | Global storage class for persistent volumes | `""` |

### Namespace

| Parameter | Description | Default |
|-----------|-------------|---------|
| `namespace.name` | Namespace name | `monorepo` |
| `namespace.labels` | Namespace labels | `{environment: production}` |

### Service Account

| Parameter | Description | Default |
|-----------|-------------|---------|
| `serviceAccount.create` | Create service account | `true` |
| `serviceAccount.name` | Service account name | `monorepo-sa` |

### PostgreSQL

| Parameter | Description | Default |
|-----------|-------------|---------|
| `postgresql.enabled` | Enable PostgreSQL | `true` |
| `postgresql.image.tag` | PostgreSQL image tag | `13` |
| `postgresql.persistence.enabled` | Enable persistence | `true` |
| `postgresql.persistence.size` | PVC size | `10Gi` |

### User Service

| Parameter | Description | Default |
|-----------|-------------|---------|
| `userService.enabled` | Enable user service | `true` |
| `userService.replicaCount` | Number of replicas | `3` |
| `userService.image.repository` | Image repository | `kevindiu/monorepo-go-example/user-service` |
| `userService.image.tag` | Image tag | `latest` |
| `userService.autoscaling.enabled` | Enable HPA | `true` |
| `userService.autoscaling.minReplicas` | Minimum replicas | `2` |
| `userService.autoscaling.maxReplicas` | Maximum replicas | `10` |

### Order Service

| Parameter | Description | Default |
|-----------|-------------|---------|
| `orderService.enabled` | Enable order service | `true` |
| `orderService.replicaCount` | Number of replicas | `3` |
| `orderService.image.repository` | Image repository | `kevindiu/monorepo-go-example/order-service` |
| `orderService.image.tag` | Image tag | `latest` |
| `orderService.autoscaling.enabled` | Enable HPA | `true` |
| `orderService.autoscaling.minReplicas` | Minimum replicas | `2` |
| `orderService.autoscaling.maxReplicas` | Maximum replicas | `10` |

### Gateway

| Parameter | Description | Default |
|-----------|-------------|---------|
| `gateway.enabled` | Enable gateway | `true` |
| `gateway.replicaCount` | Number of replicas | `3` |
| `gateway.image.repository` | Image repository | `kevindiu/monorepo-go-example/gateway` |
| `gateway.image.tag` | Image tag | `latest` |
| `gateway.service.type` | Service type | `LoadBalancer` |
| `gateway.autoscaling.enabled` | Enable HPA | `true` |
| `gateway.autoscaling.minReplicas` | Minimum replicas | `2` |
| `gateway.autoscaling.maxReplicas` | Maximum replicas | `15` |

### Ingress

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ingress.enabled` | Enable ingress | `false` |
| `ingress.className` | Ingress class name | `nginx` |
| `ingress.hosts[0].host` | Hostname | `api.monorepo.example.com` |
| `ingress.tls[0].secretName` | TLS secret name | `gateway-tls` |

## Examples

### Install with custom values

```bash
helm install my-release ./charts/monorepo-go-example \
  --set gateway.service.type=ClusterIP \
  --set ingress.enabled=true \
  --set secrets.database.password=mySecurePassword
```

### Install in a specific namespace

```bash
kubectl create namespace my-namespace
helm install my-release ./charts/monorepo-go-example -n my-namespace
```

### Using a values file

Create a `custom-values.yaml` file:

```yaml
secrets:
  database:
    password: mySecurePassword

gateway:
  service:
    type: ClusterIP
  
ingress:
  enabled: true
  hosts:
    - host: api.mydomain.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: my-tls-secret
      hosts:
        - api.mydomain.com

postgresql:
  persistence:
    size: 20Gi
    storageClass: fast-ssd
```

Install with:

```bash
helm install my-release ./charts/monorepo-go-example -f custom-values.yaml
```

### Upgrade

```bash
helm upgrade my-release ./charts/monorepo-go-example
```

### Rollback

```bash
helm rollback my-release 1
```

## Testing the Deployment

After installation, you can test the services:

```bash
# Port forward to gateway
kubectl port-forward svc/gateway 8080:80 -n monorepo

# Test user service via gateway
curl http://localhost:8080/v1/users

# Check pod status
kubectl get pods -n monorepo

# Check logs
kubectl logs -l app=user-service -n monorepo
```

## Accessing Services

### Via LoadBalancer (default)

```bash
export GATEWAY_IP=$(kubectl get svc gateway -n monorepo -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
curl http://$GATEWAY_IP/v1/users
```

### Via Port Forward

```bash
kubectl port-forward svc/gateway 8080:80 -n monorepo
curl http://localhost:8080/v1/users
```

### Via Ingress (if enabled)

```bash
curl https://api.monorepo.example.com/v1/users
```

## Monitoring

Prometheus metrics are available at `/metrics` on each service's HTTP port.

## Troubleshooting

### Check pod status

```bash
kubectl get pods -n monorepo
```

### View logs

```bash
kubectl logs -l app=user-service -n monorepo
kubectl logs -l app=order-service -n monorepo
kubectl logs -l app=gateway -n monorepo
kubectl logs -l app=postgres -n monorepo
```

### Describe resources

```bash
kubectl describe pod <pod-name> -n monorepo
kubectl describe svc gateway -n monorepo
```

### Check database connection

```bash
kubectl exec -it postgres-0 -n monorepo -- psql -U postgres -d monorepo
```

## License

Copyright (C) 2025 Kevin Diu <kevindiujp@gmail.com>

Licensed under the Apache License, Version 2.0.