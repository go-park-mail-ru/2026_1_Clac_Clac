FROM golang:1.26-alpine AS builder

# Флаг CGO_ENABLED говорит компилятору сделать
# один бинарник, без всяких СИшных библиотек
ENV GOOS=linux CGO_ENABLED=0

RUN apk add --no-cache make

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY pkg/ ./pkg/
COPY realtime/ ./realtime/

# Бинарник будет в /app/realtime/bin
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    make -C realtime build


FROM alpine:latest

# Для сертификатов и времени
RUN apk add --no-cache ca-certificates tzdata \
    && ln -sf /etc/ssl/certs/ca-certificates.crt /etc/ssl/cert.pem

WORKDIR /app

# Будем запускать НЕ из-под root
RUN adduser -D nexus && chown nexus:nexus /app

COPY --from=builder --chown=nexus:nexus /app/realtime/bin/main .

USER nexus
EXPOSE 8080
ENTRYPOINT ["./main"]
