FROM golang:1.24.1-alpine

WORKDIR /app

# Копируем структуру проекта
COPY db_service/ ./db_service/
COPY proto/ ./proto/

# Переходим в директорию db_service
WORKDIR /app/db_service

# Загружаем зависимости
RUN go mod download

# Собираем приложение
RUN go build -o db_service ./cmd/main.go

EXPOSE 50051

CMD ["./db_service"]
