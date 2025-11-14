# Build stage
FROM golang:1.24.4-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o bin/kisanlink-erp \
    cmd/server/main.go

# Runtime stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk --no-cache add ca-certificates curl

# Create non-root user
RUN addgroup -g 1000 erp && \
    adduser -D -u 1000 -G erp erp

# Set working directory
WORKDIR /home/erp

# Copy binary from builder
COPY --from=builder /app/bin/kisanlink-erp .

# Change ownership
RUN chown -R erp:erp /home/erp

# Switch to non-root user
USER erp

# Expose port
EXPOSE 8080

# Health check - uses SERVER_HTTP_PORT env var with fallback to 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=60s --retries=3 \
    CMD sh -c "curl -f http://localhost:${SERVER_HTTP_PORT:-8080}/health || exit 1"

# Run the application
CMD ["./kisanlink-erp"]
