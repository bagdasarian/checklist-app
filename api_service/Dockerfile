FROM golang:1.24.1-alpine

WORKDIR /app

# Копируем структуру проекта для replace директивы
COPY api_service/ ./api_service/
COPY db_service/ ./db_service/
COPY proto/ ./proto/

# Переходим в директорию api_service
WORKDIR /app/api_service

# Загружаем зависимости
RUN go mod download

# Собираем приложение
RUN go build -o api_service ./cmd/main.go

EXPOSE 8080 50052

CMD ["./api_service"]
