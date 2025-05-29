# Build stage
FROM golang:1.20-alpine AS builder

# Set working directory
WORKDIR /app

# Install necessary build tools
RUN apk add --no-cache gcc musl-dev

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o frigate_alerter ./cmd/frigate_alerter

# Final stage
FROM alpine:3.18

# Add timezone data
RUN apk add --no-cache tzdata ca-certificates

# Create a non-root user to run the application
RUN adduser -D -h /app appuser

# Create necessary directories with proper permissions
RUN mkdir -p /app/web/templates /app/web/static
COPY --from=builder /app/web/templates /app/web/templates
COPY --from=builder /app/web/static /app/web/static

# Copy the binary from the builder stage
COPY --from=builder /app/frigate_alerter /app/frigate_alerter

# Set working directory
WORKDIR /app

# Switch to non-root user
USER appuser

# Volume for database and config
VOLUME ["/app/data"]

# Expose the web UI port
EXPOSE 8080

# Set the entry point
ENTRYPOINT ["/app/frigate_alerter"]
