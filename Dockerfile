# E-Paper Dashboard Dockerfile
# Multi-stage build for optimized production container

# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies including CGO requirements
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO enabled for SQLite
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags='-w -s' \
    -o homeboard cmd/server/main.go

# Production stage
FROM python:3.11-alpine AS production

# Install system dependencies for Python widgets
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    gcc \
    musl-dev \
    linux-headers \
    python3-dev \
    && rm -rf /var/cache/apk/*

# Install Python dependencies for widgets
RUN pip install --no-cache-dir \
    psutil \
    pytz \
    requests \
    && apk del gcc musl-dev linux-headers python3-dev

# Create non-root user
RUN addgroup -g 1001 -S homeboard && \
    adduser -S homeboard -u 1001 -G homeboard

# Set working directory
WORKDIR /app

# Copy built binary from builder stage
COPY --from=builder /app/homeboard ./

# Copy configuration files
COPY config.json ./
COPY config-extended.json ./

# Copy widgets
COPY widgets/ ./widgets/

# Copy static assets (CSS, icons, etc.)
COPY static/ ./static/

# Copy additional files would go here if needed

# Create necessary directories
RUN mkdir -p /app/data /app/logs /app/backups && \
    chown -R homeboard:homeboard /app

# Switch to non-root user
USER homeboard

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Set environment variables
ENV GIN_MODE=release
ENV PYTHONPATH=/app/widgets
ENV CONFIG_PATH=/app/config.json

# Default command
CMD ["./homeboard", "-config", "/app/config.json", "-adk-url", "${ADK_SERVICE_URL:-http://localhost:8081}", "-verbose"]

# Labels for container metadata
LABEL maintainer="homeboard@example.com"
LABEL version="1.0.0"
LABEL description="E-Paper Dashboard - Lightweight dashboard for e-ink displays"
LABEL org.opencontainers.image.title="E-Paper Dashboard"
LABEL org.opencontainers.image.description="A comprehensive dashboard system for E-Paper displays"
LABEL org.opencontainers.image.vendor="Homeboard Project"
LABEL org.opencontainers.image.licenses="MIT"