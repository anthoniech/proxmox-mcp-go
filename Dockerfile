# Build golang app
FROM golang:1.24-alpine AS builder

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /build
COPY go.mod go.sum ./
COPY vendor vendor
COPY main.go ./
COPY app app
COPY config config
COPY server server
COPY mcp mcp

RUN CGO_ENABLED=0 GOOS=linux go build \
    -mod=vendor \
    -ldflags='-extldflags "-static" -X "main.version=v0.0.1"' \
    -o proxmox-mcp-go .

# Create non-root user
FROM alpine:3.20 AS security_provider
RUN addgroup -S nonroot && adduser -S nonroot -G nonroot
RUN mkdir /data && mkdir /logs && chown nonroot:nonroot /logs

# Final image
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /build/proxmox-mcp-go /app/
COPY --from=builder /build/config/config.yaml /app/

COPY --from=security_provider /etc/passwd /etc/passwd
COPY --from=security_provider /data /app/data
COPY --chown=nonroot:nonroot --from=security_provider /logs /app/logs
USER nonroot

WORKDIR /app
CMD ["./proxmox-mcp-go", "-v"]
