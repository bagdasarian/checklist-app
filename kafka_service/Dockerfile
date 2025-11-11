FROM golang:1.24.1-alpine

WORKDIR /app

# Копируем структуру проекта
COPY kafka_service/ ./kafka_service/
COPY proto/ ./proto/

# Переходим в директорию kafka_service
WORKDIR /app/kafka_service

# Загружаем зависимости
RUN go mod download

# Собираем приложение
RUN go build -o kafka_service ./cmd/main.go

EXPOSE 9092

CMD ["./kafka_service"]
