# Stage 1: Build
FROM docker.io/library/golang:1.25-alpine AS builder

# Accept version as build arg
ARG VERSION=dev

# Install make (required for build process)
RUN apk add --no-cache make

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build using Makefile with version injection
RUN VERSION=${VERSION} make build-local

# Stage 2: Runtime
FROM scratch

# OCI labels for metadata and GitHub integration
LABEL org.opencontainers.image.source="https://github.com/neox5/obsbox"
LABEL org.opencontainers.image.description="Telemetry signal generator for testing observability components"
LABEL org.opencontainers.image.licenses="MIT"

# Copy CA certificates for HTTPS (if needed for OTEL)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary from dist directory (build-local output location)
COPY --from=builder /build/dist/obsbox /obsbox

# Run as non-root user
USER 65534:65534

# Expose Prometheus port
EXPOSE 9090

# Default entrypoint
ENTRYPOINT ["/obsbox"]

# Default config path (override with volume mount)
CMD ["-config", "/config/config.yaml"]
