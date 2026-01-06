# Build Stage
FROM golang:1.23-alpine AS builder

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

FROM alpine:latest

WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/vigil /usr/local/bin/vigil

EXPOSE 8080
ENTRYPOINT ["vigil"]