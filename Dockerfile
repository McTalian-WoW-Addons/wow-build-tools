#checkov:skip=CKV_DOCKER_2: HEALTHCHECK is not applicable for a docker container action
#checkov:skip=CKV_DOCKER_3: USER instructions should not be used for docker container actions
# Use official Go image with specific version from go.mod
FROM golang:1.25.5-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git=2.52.0-r0

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
FROM alpine:3.23

# Install runtime dependencies
RUN apk add --no-cache \
    git=2.52.0-r0 \
    subversion=1.14.5-r1 \
    ca-certificates=20251003-r0 \
    bash=5.3.3-r1

# Copy the built binary from builder stage
COPY --from=builder /app/wow-build-tools /usr/local/bin/wow-build-tools

# Make binary executable
RUN chmod +x /usr/local/bin/wow-build-tools

# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Set entrypoint
ENTRYPOINT ["/entrypoint.sh"]
