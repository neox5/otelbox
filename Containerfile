# Stage 1: Build
FROM docker.io/library/golang:1.25-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o obsbox \
    ./cmd/obsbox

# Stage 2: Runtime
FROM scratch

# Copy CA certificates for HTTPS (if needed for OTEL)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary
COPY --from=builder /build/obsbox /obsbox

# Run as non-root user
USER 65534:65534

# Expose Prometheus port
EXPOSE 9090

# Default entrypoint
ENTRYPOINT ["/obsbox"]

# Default config path (override with volume mount)
CMD ["-config", "/config/config.yaml"]
