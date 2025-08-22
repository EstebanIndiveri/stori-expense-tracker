# Multi-stage build for minimal image size
FROM golang:1.24-alpine AS builder

# Install git and ca-certificates
RUN apk --no-cache add git ca-certificates tzdata

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o api ./cmd/api

# Final stage - minimal image
FROM scratch

# Copy ca-certificates from builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /app/api /api

# Expose port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["/api"]
