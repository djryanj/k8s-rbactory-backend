# Global args
ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ARG BUILD_DATE
ARG VCS_REF
ARG VERSION=0.0.1

# Builder stage
FROM --platform=${BUILDPLATFORM} golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

# Local args
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    GOARM=${TARGETVARIANT#v} \
    go build \
    -a \
    -installsuffix cgo \
    -ldflags="-w -s -extldflags '-static' -X main.Version=${VERSION:-dev} -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o api \
    ./cmd/api

RUN file api && \
    chmod +x api

# final stage
FROM gcr.io/distroless/static-debian13:nonroot

ARG BUILD_DATE
ARG VCS_REF
ARG VERSION=0.0.1

LABEL org.opencontainers.image.created="${BUILD_DATE}" \
    org.opencontainers.image.title="K8s RBACtory Backend" \
    org.opencontainers.image.description="Backend API for K8s RBACtory - RBAC policy builder" \
    org.opencontainers.image.licenses="Apache-2.0" \
    org.opencontainers.image.version="${VERSION}" \
    org.opencontainers.image.revision="${VCS_REF}" \
    org.opencontainers.image.source="https://github.com/djryanj/k8s-rbactory-backend" \
    org.opencontainers.image.documentation="https://github.com/djryanj/k8s-rbactory-backend/blob/main/README.md"

COPY --from=builder /app/api /api

EXPOSE 8080

ENTRYPOINT ["/api"]