# Этап 1: Сборка (Builder)
FROM golang:1.22-alpine AS builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы зависимостей и скачиваем их
# Это делается до копирования кода, чтобы кэшировать слои Docker
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Собираем бинарник.
# CGO_ENABLED=0 делает бинарник статическим (необходим для запуска в чистом alpine)
RUN CGO_ENABLED=0 GOOS=linux go build -o user-service ./cmd/main.go

# Этап 2: Финальный легковесный образ (Runner)
FROM alpine:latest

WORKDIR /app

# Копируем скомпилированный бинарник из первого этапа
COPY --from=builder /app/user-service .

# Копируем конфиги (обязательно!)
COPY --from=builder /app/config.yaml .
# Если используешь .env напрямую, раскомментируй строку ниже (но лучше прокидывать через docker-compose)
# COPY --from=builder /app/.env .

# Открываем порт, на котором работает gRPC
EXPOSE 50052

# Команда для запуска
CMD ["./user-service"]
