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

## Installation

### Local Development

1. Clone the repository:

```bash
git clone https://github.com/djryanj/k8s-rbactory-backend.git
cd k8s-rbactory-backend/backend
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
  -v ~/.kube/config:/root/.kube/config:ro \
  rbac-generator-api:latest
```

### Kubernetes Deployment

1. Apply RBAC configuration:

```bash
kubectl apply -f deploy/rbac.yaml
```

2. Deploy the application:

```bash
kubectl apply -f deploy/deployment.yaml
```

3. Verify the deployment:

```bash
kubectl get pods -l app=rbac-generator-api
kubectl logs -l app=rbac-generator-api
```

## Configuration

The application is configured using environment variables:

| Variable          | Description                                   | Default                 | Required |
| ----------------- | --------------------------------------------- | ----------------------- | -------- |
| `PORT`            | HTTP server port                              | `8080`                  | No       |
| `ALLOWED_ORIGINS` | CORS allowed origins (comma-separated)        | `http://localhost:3000` | No       |
| `KUBECONFIG`      | Path to kubeconfig file (out-of-cluster only) | `~/.kube/config`        | No       |

### Example Configuration

```bash
export PORT=9090
export ALLOWED_ORIGINS="https://app.example.com,https://admin.example.com"
./api
```

## API Documentation

API documentation is available in swagger format at [api/swagger.json](api/swagger.json).

When the server is running, it is browsable at `/swagger` (e.g., http://\<base-url\>/swagger)

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

## Deployment

### RBAC Permissions

The application requires the following Kubernetes RBAC permissions:

```yaml
rules:
  # Read RBAC resources
  - apiGroups: ["rbac.authorization.k8s.io"]
    resources:
      - roles
      - clusterroles
      - rolebindings
      - clusterrolebindings
    verbs: ["get", "list", "watch"]

  # Read namespaces
  - apiGroups: [""]
    resources:
      - namespaces
    verbs: ["get", "list"]

  # Read cluster version
  - nonResourceURLs: ["/version"]
    verbs: ["get"]
```

### Security Considerations

#### In-Cluster Deployment

1. **Service Account**: Use a dedicated service account with minimal permissions
2. **Network Policies**: Restrict ingress/egress traffic
3. **Pod Security**: Run as non-root user with read-only filesystem
4. **Resource Limits**: Set appropriate CPU and memory limits

#### CORS Configuration

Configure allowed origins for production:

```yaml
env:
  - name: ALLOWED_ORIGINS
    value: "https://your-frontend.example.com"
```

#### TLS/HTTPS

For production deployments, use an ingress controller with TLS:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: rbac-generator-api
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
    - hosts:
        - api.example.com
      secretName: rbac-api-tls
  rules:
    - host: api.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: rbac-generator-api
                port:
                  number: 80
```

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
kubectl logs -l app=rbac-generator-api --tail=100 -f

# View logs for specific pod
kubectl logs rbac-generator-api-xxx-yyy -f

# View previous container logs
kubectl logs rbac-generator-api-xxx-yyy --previous
```

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Run linter (`make lint`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Write tests for new features
- Update documentation as needed
- Keep commits atomic and well-described

### Testing Requirements

All pull requests must:

- Pass all existing tests
- Include tests for new functionality
- Maintain or improve code coverage
- Pass linter checks

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
