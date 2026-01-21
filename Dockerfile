# Multi-stage build for Term Idle
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the SSH server
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o ssh-server cmd/ssh-server/main.go

# Build the game client
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o term-idle cmd/term-idle/main.go

# Final runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata sqlite

# Create app user
RUN addgroup -g 1001 -S termidle && \
    adduser -u 1001 -S termidle -G termidle

# Set working directory
WORKDIR /app

# Create necessary directories
RUN mkdir -p /app/data /app/logs /app/configs && \
    chown -R termidle:termidle /app

# Copy binaries from builder stage
COPY --from=builder /app/ssh-server /app/ssh-server
COPY --from=builder /app/term-idle /app/term-idle

# Copy configuration files
COPY configs/config.yaml /app/configs/config.yaml

# Set ownership
RUN chmod +x /app/ssh-server /app/term-idle && \
    chown termidle:termidle /app/ssh-server /app/term-idle

# Switch to non-root user
USER termidle

# Expose ports
EXPOSE 2222 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD nc -z localhost 8080 || exit 1

# Default command
CMD ["/app/ssh-server"]