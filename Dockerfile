FROM golang:1.25-alpine AS builder
WORKDIR /app
RUN --mount=type=cache,target=/go/pkg/mod \ --mount=type=cache,target=/root/.cache/go-build \ go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \ --mount=type=cache,target=/root/.cache/go-build \ CGO_ENABLED=0 GOOS=linux go build -o gateway-vps ./cmd/gateway-vps

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata \ && addgroup -g 1001 -S appuser \ && adduser -S -D -u 1001 -G appuser 
COPY --from=builder /gateway-vps /gateway-vps 
USER 1001
EXPOSE 8080
ENTRYPOINT [ "/gateway-vps" ]

