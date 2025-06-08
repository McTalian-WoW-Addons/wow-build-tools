# Use official Go image with specific version from go.mod
FROM golang:1.23.6-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN go build -o wow-build-tools .

# Final stage - use alpine for smaller image size
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache \
    git \
    subversion \
    ca-certificates \
    bash

# Create a non-root user
RUN adduser -D -s /bin/bash github

# Create cache directories with proper permissions
RUN mkdir -p /home/github/.wow-build-tools/cache/externals && \
    chown -R github:github /home/github/.wow-build-tools

# Copy the built binary from builder stage
COPY --from=builder /app/wow-build-tools /usr/local/bin/wow-build-tools

# Make binary executable
RUN chmod +x /usr/local/bin/wow-build-tools

# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Set user
USER github

# Set entrypoint
ENTRYPOINT ["/entrypoint.sh"]
