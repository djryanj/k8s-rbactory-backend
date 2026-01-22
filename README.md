# K8s RBACtory Backend API

A high-performance REST API service for retrieving Kubernetes Role-Based Access Control (RBAC) configurations for analysis alongside K8s RBACtory Frontend.

This API service exists primarily because of the limitations inherent in using `kubectl proxy` (i.e., CORS and least privilege) to handle requests from the frontend.

This API service can run in-cluster (recommended), or it can be run locally and leverage `kubeconfig` to connect to the cluster.

**_NOTE_**: This is a READ-ONLY tool. It is **NOT** a goal of this project to implement the ability to write RBAC configurations of any kind into the cluster.

## Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
  - [Local Development](#local-development)
  - [Docker](#docker)
  - [Kubernetes Deployment](#kubernetes-deployment)
- [Configuration](#configuration)
- [API Documentation](#api-documentation)
- [Development](#development)
  - [Building](#building)
  - [Testing](#testing)
  - [Code Quality](#code-quality)
- [Deployment](#deployment)
  - [RBAC Permissions](#rbac-permissions)
  - [Security Considerations](#security-considerations)
- [Monitoring and Observability](#monitoring-and-observability)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

## Features

- **Comprehensive RBAC Analysis**: Query roles, cluster roles, role bindings, and cluster role bindings
- **Principal Discovery**: Identify all users, groups, and service accounts with their associated permissions
- **Relationship Mapping**: Understand connections between roles, bindings, and principals
- **Pagination Support**: Efficient handling of large RBAC configurations
- **Resource Counting**: Quick overview of RBAC resource distribution
- **In-Cluster and Remote Access**: Works both inside Kubernetes clusters and with external kubeconfig
- **High Performance**: Built with Go for optimal speed and resource efficiency
- **Production Ready**: Includes rate limiting, request tracing, compression, and graceful shutdown
- **Comprehensive Logging**: Structured logging with request correlation

## Architecture

The application follows a clean architecture pattern with clear separation of concerns:

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────┐
│      Middleware Layer           │
│  - Request ID                   │
│  - Logging                      │
│  - Rate Limiting                │
│  - CORS                         │
│  - Compression                  │
│  - Timeout                      │
│  - Recovery                     │
└──────────┬──────────────────────┘
           │
           ▼
┌─────────────────────────────────┐
│      Handler Layer              │
│  - RBAC Handlers                │
│  - Cluster Handlers             │
│  - Input Validation             │
└──────────┬──────────────────────┘
           │
           ▼
┌─────────────────────────────────┐
│   Kubernetes Client Layer       │
│  - Client Interface             │
│  - RBAC Operations              │
│  - Resource Conversion          │
└──────────┬──────────────────────┘
           │
           ▼
┌─────────────────────────────────┐
│    Kubernetes API Server        │
└─────────────────────────────────┘
```

## Prerequisites

### For Local Development

- Go 1.21 or higher
- Access to a Kubernetes cluster
- kubectl configured with appropriate credentials

### For Docker Deployment

- Docker 20.10 or higher
- Docker Compose (optional)

### For Kubernetes Deployment

- Kubernetes 1.24 or higher
- kubectl access with cluster-admin privileges (for initial setup)

## Deployment

### Local Development

1. Clone the repository:

```bash
git clone https://github.com/djryanj/k8s-rbactory-backend.git
cd k8s-rbactory-backend
```

2. Install dependencies:

```bash
go mod download
```

3. Build the application:

```bash
make build
```

4. Run the application:

```bash
make run
```

The API will be available at `http://localhost:8080`.

### Docker

1. Build the Docker image:

```bash
make docker-build
```

2. Run the container:

```bash
docker run -p 8080:8080 \
  -v ~/.kube/config:/home/nonroot/.kube/config:ro \
  k8s-rbactory-backend:latest
```

### Kubernetes Deployment

Deployment manifests are available in the [hack/k8s-manfiests](./hack/k8s-manifests/) directory. A `kustomization.yaml` file is provided for use with kustomize (recommended).

#### Method 1: Direct from GitHub (Recommended for Quick Testing)

Deploy directly from the GitHub repository without cloning:

```bash
kubectl apply -k github.com/djryanj/k8s-rbactory-backend/hack/k8s-manifests
```

#### Method 2: From Local Clone

Clone the repository and deploy:

```bash
# Clone the repository
git clone https://github.com/djryanj/k8s-rbactory-backend.git
cd k8s-rbactory-backend

# Deploy
kubectl apply -k hack/k8s-manifests
```

#### Method 3: Using Kustomize CLI

For more control and to preview changes:

```bash
# Preview what will be deployed
kustomize build hack/k8s-manifests

# Deploy using kustomize
kustomize build hack/k8s-manifests | kubectl apply -f -
```

#### Method 4: Customize deployment using your own overlay

Create a `kustomization.yaml` file that extends what's in GitHub:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - github.com/djryanj/k8s-rbactory-backend/hack/k8s-manifests

# Override namespace
namespace: my-custom-namespace

# Add custom labels
commonLabels:
  team: my-team
  cost-center: "12345"

# Override image
images:
  - name: k8s-rbactory-backend
    newName: my-registry.example.com/k8s-rbactory-backend
    newTag: v2.0.0

# Override replicas
replicas:
  - name: k8s-rbactory-backend
    count: 5

# patch the backend for CORS
patches:
  - target:
      kind: Deployment
      name: k8s-rbactory-backend
    patch: |-
      - op: add
        path: /spec/template/spec/containers/0/env/-
        value:
          name: ALLOWED_ORIGINS
          value: "http://localhost:5173"
```

Deploy that:

```shell
kubectl apply -k kustomization.yaml
```

#### Verify the deployment:

```bash
kubectl get pods -n k8s-rbactory -l app=k8s-rbactory-backend
kubectl logs -n k8s-rbactory -l app=k8s-rbactory-backend
```

### Ingress

A reference [ingress manifest](./hack/k8s-manifests/ingress.yaml) is provided in [hack/k8s-manifests](./hack/k8s-manifests/) for reference but it is **NOT** included in the provided `kustomization.yaml`.

### Burstable By Default

The [provided manifests](./hack/k8s-manifests/) deliberately put this deployment in the [burstable QoS class](https://kubernetes.io/docs/concepts/workloads/pods/pod-qos/#burstable) and some basic tolerations for spot instances. This is done under the assumption that this deployment is non-critical to most clusters.

## Security Considerations

### DO NOT EXPOSE PUBLICLY

This service WILL expose information about your cluster(s) to unauthenticated users that could potentially be used in a malicious way. Therefore, **_DO NOT EXPOSE IT ON A PUBLICLY REACHABLE INGRESS/URL_**.

The provided [ingress manifest](./hack/k8s-manifests/ingress.yaml) deliberately uses a nonstandard `ingressClassName` to help avoid accidentally exposing it on a default ingress.

If your cluster is only reachable via a public Ingress, then use:

```shell
kubectl port-forward -n k8s-rbactory service/k8s-rbactory-backend 8080 8080
```

And update the API URL parameter in the frontend to `http://localhost:8080/v1/api`.

### Service Account with Proper RBAC

It would be ironic to use an RBAC tool without proper RBAC in place. So don't do that.

The [provided manifests](./hack/k8s-manifests/) create and use a service account as well as minimal RBAC required for this deployment to get what's needed from the cluster in a read-only manner. Use them, or something like them, please.

### CORS Configuration

Configure allowed origins for production:

```yaml
env:
  - name: ALLOWED_ORIGINS
    value: "https://your-frontend.example.com"
```

### TLS/HTTPS

For production deployments, use an ingress controller with TLS. See the [sample `ingress.yaml`](./hack/k8s-manifests/ingress.yaml).

## Configuration

The application is configured using environment variables:

| Variable          | Description                                   | Default                 | Required |
| ----------------- | --------------------------------------------- | ----------------------- | -------- |
| `PORT`            | HTTP server port                              | `8080`                  | No       |
| `ALLOWED_ORIGINS` | CORS allowed origins (comma-separated)        | `http://localhost:5713` | No       |
| `KUBECONFIG`      | Path to kubeconfig file (out-of-cluster only) | `~/.kube/config`        | No       |

### Example Configuration

```bash
export PORT=9090
export ALLOWED_ORIGINS="https://app.example.com,https://admin.example.com"
./api
```

## API Documentation

API documentation is available in swagger format at [api/swagger.json](api/swagger.json).

When the server is running, it is browsable at `/swagger` (e.g., https://\<base-url\>/swagger)

### Rate Limiting

The API implements rate limiting to prevent abuse:

- **Default**: 100 requests per second per IP address
- **Burst**: 200 requests
- **Response**: HTTP 429 when limit exceeded

### Request Tracing

All requests include an `X-Request-ID` header for tracing:

```http
X-Request-ID: 550e8400-e29b-41d4-a716-446655440000
```

If the client provides this header, it will be preserved; otherwise, a new UUID is generated.

## Development

### Building

```bash
# Build the binary
make build

# Build Docker image
make docker-build

# Build and run
make dev
```

### Testing

The project includes comprehensive test coverage:

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests
make test-integration

# Run tests with coverage
make test-coverage

# Run tests with race detector
make test-race

# Run benchmarks
make test-bench

# Run specific test package
make test-handlers
make test-middleware
make test-k8s
```

#### Test Coverage

The project maintains high test coverage across all packages:

```bash
# Generate coverage report
make test-coverage

# View coverage summary
make test-summary
```

Current coverage targets:

- Overall: >70%
- Handlers: >80%
- Middleware: >85%
- Kubernetes client: >75%

### Code Quality

```bash
# Run linter
make lint

# Format code
make fmt

# Run go vet
make vet

# Run all quality checks
make check

# Pre-commit checks (fast)
make pre-commit

# Full CI pipeline
make ci
```

### Development Tools

Install required development tools:

```bash
make install-tools
```

This installs:

- golangci-lint
- goimports
- staticcheck

## Monitoring and Observability

### Logging

The application uses structured JSON logging with the following fields:

- `time`: ISO 8601 timestamp
- `level`: Log level (DEBUG, INFO, WARN, ERROR)
- `msg`: Log message
- `request_id`: Unique request identifier
- `method`: HTTP method
- `path`: Request path
- `status`: HTTP status code
- `duration`: Request duration
- Additional context-specific fields

Example log output:

```json
{
  "time": "2024-01-06T15:30:45Z",
  "level": "INFO",
  "msg": "http request completed",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "method": "GET",
  "path": "/api/v1/roles",
  "status": 200,
  "duration": "150ms",
  "duration_ms": 150,
  "remote_addr": "192.168.1.100:54321",
  "user_agent": "Mozilla/5.0..."
}
```

### Health Checks

The application provides health check endpoints for Kubernetes probes:

```yaml
livenessProbe:
  httpGet:
    path: /api/v1/healthz
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 30

readinessProbe:
  httpGet:
    path: /api/v1/healthz
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
```

### Metrics

At this time, no metrics are exposed as they are not deemed important enough for a small project like this. Feel free to contribute that functionality if you feel it's needed.

## Troubleshooting

### Common Issues

#### Connection Refused

**Problem**: Cannot connect to Kubernetes API server

**Solution**:

```bash
# Verify kubeconfig
kubectl cluster-info

# Check service account (in-cluster)
kubectl get serviceaccount rbac-generator -n default

# Check RBAC permissions
kubectl auth can-i list roles --as=system:serviceaccount:default:rbac-generator
```

#### Permission Denied

**Problem**: Insufficient RBAC permissions

**Solution**:

```bash
# Verify ClusterRoleBinding
kubectl get clusterrolebinding rbac-generator-reader-binding

# Check what the service account can do
kubectl auth can-i --list --as=system:serviceaccount:default:rbac-generator
```

#### Rate Limiting

**Problem**: Receiving 429 Too Many Requests

**Solution**: Implement client-side rate limiting or request the rate limit to be increased.

#### Timeout Errors

**Problem**: Requests timing out

**Solution**:

- Check cluster performance
- Verify network connectivity
- Increase timeout configuration if needed
- Use pagination for large result sets

### Debug Mode

Enable debug logging:

```bash
# Set log level to debug
export LOG_LEVEL=debug
./api
```

### Viewing Logs

```bash
# View application logs
kubectl logs -l app=k8s-rbactory-backend --tail=100 -f

# View logs for specific pod
kubectl logs k8s-rbactory-backend-xxx-yyy -f

# View previous container logs
kubectl logs k8s-rbactory-backend-xxx-yyy --previous
```

## Code of Conduct

This project adheres to a Code of Conduct that all contributors are expected to follow. By participating, you are expected to uphold this code. Please report unacceptable behavior to [INSERT EMAIL ADDRESS].

Please read our full [Code of Conduct](CODE_OF_CONDUCT.md) before contributing.

## Contributing

See [CONTRIBUTING](CONTRIBUTING.md).

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Support

For issues, questions, or contributions:

- **Issues**: [GitHub Issues](https://github.com/djryanj/k8s-rbactory-backend/issues)
- **Discussions**: [GitHub Discussions](https://github.com/djryanj/k8s-rbactory-backend/discussions)
- **Documentation**: [Wiki](https://github.com/djryanj/k8s-rbactory-backend/wiki)

## Acknowledgments

- Built with [client-go](https://github.com/kubernetes/client-go)
- Uses [Gorilla Mux](https://github.com/gorilla/mux) for routing
- Inspired by Kubernetes RBAC best practices

---

**Note**: This is a read-only tool that does not modify any Kubernetes resources. It requires only read permissions on RBAC resources.
