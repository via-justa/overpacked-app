# Build stage
FROM golang:1.26.4-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api ./cmd/api

# Runtime stage
FROM alpine:3.22

RUN apk --no-cache add ca-certificates \
    && addgroup -S app && adduser -S app -G app

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/api .

# Copy any static assets if needed
COPY --from=builder /app/internal/migrations/sql ./internal/migrations/sql
COPY --from=builder /app/internal/seeds/data ./internal/seeds/data

RUN chown -R app:app /app && \

# Pre-create the default backup directory owned by the runtime user. A fresh
# Docker named volume mounted here inherits this ownership, so the non-root
# process can write backups. (On Kubernetes the PVC is made writable via fsGroup.)
  mkdir -p /workspace/.backups && chown -R app:app /workspace

USER app

EXPOSE 8000

CMD ["./api"]
