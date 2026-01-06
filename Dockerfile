# Build Stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy dependencies first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build Unified Binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/vigil ./cmd/vigil

# --- Final Image ---
FROM alpine:latest

WORKDIR /app
# Install dependencies for both roles:
# - Controller: ca-certificates, tzdata
# - Agent: shutdown
RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/vigil /usr/local/bin/vigil

# Default Entrypoint
ENTRYPOINT ["vigil"]
CMD ["--help"]