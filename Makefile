# Переменные
PROTO_PATH = proto

# Основные команды для разработки
.PHONY: up
up:
	@echo "Запуск всех сервисов..."
	docker-compose up --build

.PHONY: up-d
up-d:
	@echo "Запуск всех сервисов в фоновом режиме..."
	docker-compose up --build -d

.PHONY: down
down:
	@echo "Остановка всех сервисов..."
	docker-compose down

.PHONY: restart
restart: down up-d
	@echo "Сервисы перезапущены"

.PHONY: logs
logs:
	@echo "Просмотр логов всех сервисов..."
	docker-compose logs -f

.PHONY: status
status:
	@echo "Статус сервисов..."
	docker-compose ps

# Очистка
.PHONY: clean
clean:
	@echo "Очистка Docker контейнеров и volumes..."
	docker-compose down -v

.PHONY: clean-all
clean-all:
	@echo "Полная очистка Docker (контейнеры, volumes, образы)..."
	docker-compose down -v --rmi all
	@echo "Очистка завершена"

# Генерация proto файлов
.PHONY: proto
proto: proto-db proto-api proto-kafka
	@echo "Генерация proto файлов завершена!"

.PHONY: proto-db
proto-db:
	@echo "Генерация gRPC для DB Service..."
	protoc --go_out=./db_service/pkg/pb \
		--go-grpc_out=./db_service/pkg/pb \
		--proto_path=$(PROTO_PATH) \
		db_service.proto
	@echo "DB Service proto сгенерирован"

.PHONY: proto-api
proto-api:
	@echo "Генерация gRPC Gateway для API Service..."
	protoc -I . \
		--go_out=./api_service/pkg/pb \
		--go-grpc_out=./api_service/pkg/pb \
		--grpc-gateway_out=./api_service/pkg/pb \
		--openapiv2_out=./api_service/swagger \
		--proto_path=$(PROTO_PATH) \
		api_service.proto
	@echo "API Service proto + Swagger сгенерированы"

.PHONY: proto-kafka
proto-kafka:
	@echo "Генерация Kafka proto..."
	protoc --go_out=./kafka_service/pkg/pb \
		--proto_path=$(PROTO_PATH) \
		kafka_service.proto
	protoc --go_out=./api_service/pkg/pb/kafka \
		--proto_path=$(PROTO_PATH) \
		kafka_service.proto
	@echo "Kafka proto сгенерирован"

# Миграции базы данных
.PHONY: migrate-up
migrate-up:
	@echo "Применение миграций..."
	docker-compose exec postgres psql -U docker -d test_db -f /docker-entrypoint-initdb.d/000001_init_schema.up.sql

.PHONY: migrate-up-docker
migrate-up-docker:
	@echo "Применение миграций через migrate tool..."
	docker run -v $(PWD)/db_service/migrations:/migrations --network host migrate/migrate \
		-path=/migrations \
		-database "postgres://docker:docker@localhost:5433/test_db?sslmode=disable" \
		up

# Зависимости
.PHONY: deps
deps:
	@echo "Обновление зависимостей..."
	cd api_service && go mod tidy
	cd db_service && go mod tidy
	cd kafka_service && go mod tidy
	@echo "Зависимости обновлены"

# Подключение к сервисам
.PHONY: psql
psql:
	@echo "Подключение к PostgreSQL..."
	docker-compose exec postgres psql -U docker -d test_db

.PHONY: redis-cli
redis-cli:
	@echo "Подключение к Redis..."
	docker-compose exec redis redis-cli

.PHONY: redis-monitor
redis-monitor:
	@echo "Мониторинг Redis команд..."
	docker-compose exec redis redis-cli MONITOR

.PHONY: redis-keys
redis-keys:
	@echo "Все ключи в Redis..."
	docker-compose exec redis redis-cli KEYS "*"

.PHONY: redis-flush
redis-flush:
	@echo "Очистка Redis кэша..."
	docker-compose exec redis redis-cli FLUSHDB
	@echo "Redis кэш очищен"

.PHONY: kafka-topics
kafka-topics:
	@echo "Список топиков Kafka..."
	docker-compose exec kafka kafka-topics --list --bootstrap-server localhost:9092



# Помощь
.PHONY: help
help:
	@echo "════════════════════════════════════════════════════════════"
	@echo "  Checklist Application - Доступные команды"
	@echo "════════════════════════════════════════════════════════════"
	@echo ""
	@echo "РАЗРАБОТКА:"
	@echo "  make up              - Запуск всех сервисов"
	@echo "  make up-d            - Запуск в фоновом режиме"
	@echo "  make down            - Остановка всех сервисов"
	@echo "  make restart         - Перезапуск сервисов"
	@echo "  make logs            - Просмотр логов"
	@echo "  make status          - Статус контейнеров"
	@echo ""
	@echo "PROTO ФАЙЛЫ:"
	@echo "  make proto           - Генерация всех proto файлов"
	@echo "  make proto-db        - Генерация DB Service proto"
	@echo "  make proto-api       - Генерация API Service proto + Swagger"
	@echo "  make proto-kafka     - Генерация Kafka proto"
	@echo ""
	@echo "БАЗА ДАННЫХ:"
	@echo "  make migrate-up      - Применить миграции"
	@echo "  make migrate-up-docker - Применить миграции через migrate tool"
	@echo "  make psql            - Подключиться к PostgreSQL"
	@echo ""
	@echo "REDIS:"
	@echo "  make redis-cli       - Подключиться к Redis CLI"
	@echo "  make redis-monitor   - Мониторинг Redis команд"
	@echo "  make redis-keys      - Показать все ключи"
	@echo "  make redis-flush     - Очистить Redis кэш"
	@echo ""
	@echo "KAFKA:"
	@echo "  make kafka-topics    - Список топиков Kafka"
	@echo ""
	@echo "ЗАВИСИМОСТИ:"
	@echo "  make deps            - Обновить go.mod для всех сервисов"
	@echo ""
	@echo "ОЧИСТКА:"
	@echo "  make clean           - Остановить и удалить volumes"
	@echo "  make clean-all       - Полная очистка (включая образы)"
	@echo ""
	@echo "ПОМОЩЬ:"
	@echo "  make help            - Показать эту справку"
	@echo ""
	@echo "════════════════════════════════════════════════════════════"

.DEFAULT_GOAL := help