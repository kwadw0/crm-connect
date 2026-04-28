# ---- Build Stage ----
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install dependencies for build (optional but common)
RUN apk add --no-cache git

# Copy go mod files first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary (static, Linux-compatible)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd

# ---- Runtime Stage ----
FROM alpine:latest

WORKDIR /app

# Install CA certificates (needed for HTTPS calls)
RUN apk add --no-cache ca-certificates

# Create a non-root user
RUN adduser -D crmuser

# Copy binary from builder and set ownership
COPY --from=builder --chown=crmuser:crmuser /app/main .

USER crmuser

# Expose your API port (application uses 3000)
EXPOSE 3000

# Run the binary
CMD ["./main"]