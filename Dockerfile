# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies (gcc and musl-dev needed for CGO/SQLite)
RUN apk add --no-cache git make gcc musl-dev

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code (excluding test files via .dockerignore)
COPY . .

# Generate code with sqlc
RUN go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate

# Build the application with optimizations
# -trimpath removes file system paths from the executable
# -ldflags="-s -w" strips debug information to reduce binary size
ARG VERSION=dev
ARG BUILD_TIME
RUN CGO_ENABLED=1 \
    go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}" \
    -o web \
    ./cmd/web

# Final stage - minimal runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/web .

# Copy UI assets and migrations
COPY --from=builder /build/ui ./ui
COPY --from=builder /build/migrations ./migrations

# Create directory for database with proper permissions
RUN mkdir -p /data && chown -R appuser:appuser /data /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Set default environment variables
ENV ADDR=":8080"
ENV DSN="/data/freelance_tracker.db"

# Run the application
CMD ["./web", "-addr=:8080", "-dsn=/data/freelance_tracker.db"]
