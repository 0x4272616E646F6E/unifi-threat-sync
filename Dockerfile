FROM golang:1.25.2-alpine AS builder

RUN apk add --no-cache \
    ca-certificates \
    git \
    tzdata

RUN adduser -D -u 10001 appuser

WORKDIR /build

COPY go.mod go.sum* ./

RUN go mod download && go mod verify

COPY . .

# Build with security flags and optimizations
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -ldflags="-w -s -extldflags '-static'" \
    -trimpath \
    -a \
    -o /build/unifi-threat-sync \
    ./cmd/unifi-threat-sync

FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder --chown=nonroot:nonroot /build/unifi-threat-sync /app/unifi-threat-sync

WORKDIR /app

USER nonroot:nonroot

# Expose health check port
EXPOSE 8080

# Note: Distroless doesn't include shell/curl, so health checks
# should be done externally (e.g., Kubernetes probes, Docker Compose healthcheck with external curl)
# For production, use: curl -f http://localhost:8080/health || exit 1

ENV TZ=UTC

ENTRYPOINT ["/app/unifi-threat-sync"]
