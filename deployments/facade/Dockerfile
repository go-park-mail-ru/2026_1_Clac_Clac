FROM golang:1.26-alpine AS builder

ENV GOOS=linux CGO_ENABLED=0
RUN apk add --no-cache make git

WORKDIR /app

# Скачиваем grpcurl прямо через Go
# Он скомпилируется под ту же архитектуру, что и твое приложение
RUN go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY pkg/ ./pkg/
COPY facade/ ./facade/

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    make -C facade build


FROM alpine:latest

# Устанавливаем только стандартные утилиты, которые точно есть
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    bind-tools \
    iproute2 \
    curl \
    busybox-extras \
    && ln -sf /etc/ssl/certs/ca-certificates.crt /etc/ssl/cert.pem

WORKDIR /app

RUN adduser -D nexus && chown nexus:nexus /app

# Копируем твой бинарник
COPY --from=builder --chown=nexus:nexus /app/facade/bin/main .

# Копируем grpcurl, который мы собрали в builder
COPY --from=builder /go/bin/grpcurl /usr/local/bin/grpcurl

USER nexus
EXPOSE 8080
ENTRYPOINT ["./main"]
