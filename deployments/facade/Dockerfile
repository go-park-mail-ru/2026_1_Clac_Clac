FROM golang:1.26-alpine AS builder

ENV GOOS=linux CGO_ENABLED=0
RUN apk add --no-cache make

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY pkg/ ./pkg/
COPY facade/ ./facade/

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    make -C facade build


FROM alpine:latest

# 1. Устанавливаем базу + инструменты отладки
# grpcurl - для gRPC запросов
# bind-tools - для проверки DNS (dig, nslookup)
# iproute2 - для ss и продвинутой работы с сетью
# curl - для проверки HTTP/2
RUN apk add --no-cache \
    --repository=https://dl-cdn.alpinelinux.org/alpine/edge/community \
    ca-certificates \
    tzdata \
    grpcurl \
    bind-tools \
    iproute2 \
    curl \
    busybox-extras \
    && ln -sf /etc/ssl/certs/ca-certificates.crt /etc/ssl/cert.pem

WORKDIR /app

RUN adduser -D nexus && chown nexus:nexus /app

COPY --from=builder --chown=nexus:nexus /app/facade/bin/main .

USER nexus
EXPOSE 8080
ENTRYPOINT ["./main"]
