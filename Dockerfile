# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main cmd/main.go

# Build the migration tool
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o migrate cmd/migrate/main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache sqlite ca-certificates tzdata

# Create app user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Create data directory
RUN mkdir -p /data && chown appuser:appgroup /data

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/migrate .

# Change to non-root user
USER appuser

# Expose port (if needed for future web interface)
EXPOSE 8080

# Set environment variables
ENV DATA_PATH=/data
ENV RUN_MODE=scheduler
ENV RUN_AT_STARTUP=true

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD [ -f /data/cine_pulse.db ] || exit 1

# Run the application
CMD ["./main"]