
FROM golang:1.25.0-alpine3.22 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /app/gateway-vps ./cmd/gateway-vps

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata \
    && addgroup -g 1001 -S appuser \
    && adduser -S -D -u 1001 -G appuser appuser

COPY --from=builder --chown=appuser:appuser /app/gateway-vps /gateway-vps

ARG SOURCE=https://github.com/karma-234/gateway-vps
LABEL org.opencontainers.image.name="gateway-vps" \
    org.opencontainers.image.description="A Go-based ISO 8583 gateway for Fineract" \
    org.opencontainers.image.licenses="MIT" \
    org.opencontainers.image.source="${SOURCE}"

USER 1001
EXPOSE 8080 8443
ENTRYPOINT ["/gateway-vps"]