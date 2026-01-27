# Stage 1: Build Nuxt static files
FROM node:20-alpine AS web-builder

WORKDIR /web

# Copy package files
COPY web/package*.json ./

# Install dependencies
RUN npm ci

# Cache bust - change this value to force rebuild
ARG CACHE_BUST=3

# Copy web source (excluding .nuxt and .output via .dockerignore)
COPY web/ ./

# Build static files
RUN npm run generate

# Stage 2: Build Go binary
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Copy built static files from web-builder
COPY --from=web-builder /web/.output/public ./internal/api/rest/static/

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o inceptor ./cmd/inceptor

# Stage 3: Production image
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS requests and tzdata for timezones
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' appuser

# Create data directories
RUN mkdir -p /app/data/crashes /app/configs && chown -R appuser:appuser /app

# Copy binary from builder
COPY --from=builder /app/inceptor /app/inceptor

# Copy default config
COPY configs/config.example.yaml /app/configs/config.yaml

# Switch to non-root user
USER appuser

# Expose ports
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Set environment variables
ENV INCEPTOR_SERVER_HOST=0.0.0.0 \
    INCEPTOR_SERVER_REST_PORT=8080 \
    INCEPTOR_SERVER_GRPC_PORT=9090 \
    INCEPTOR_STORAGE_SQLITE_PATH=/app/data/inceptor.db \
    INCEPTOR_STORAGE_LOGS_PATH=/app/data/crashes

# Run the application
CMD ["/app/inceptor"]
