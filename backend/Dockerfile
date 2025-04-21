# Stage 1: Build the Go application
FROM golang:1.22-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum first for efficient caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o server

# Stage 2: Create the final image (with minimal size)
FROM alpine:latest

# Install CA certificates for HTTPS calls
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/server /app/

# Create a non-root user and group
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
# Switch to the non-root user
USER appuser

# Expose the port used by the application
EXPOSE 8080

# Set the entry point
ENTRYPOINT ["/app/server"] 