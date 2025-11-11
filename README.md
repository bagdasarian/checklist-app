# Checklist Application

Микросервисное приложение для управления задачами (checklist) с использованием Go, gRPC, PostgreSQL, Redis и Kafka.

## Архитектура

- **api_service**: HTTP Gateway (gRPC Gateway) и gRPC сервер с JWT аутентификацией
- **db_service**: gRPC сервис для работы с базой данных (PostgreSQL + Redis)
- **kafka_service**: Сервис для обработки событий из Kafka

**Технологии:**
- **gRPC Gateway** - преобразует HTTP/REST запросы в gRPC вызовы
- Все HTTP endpoints автоматически транслируются в gRPC методы через gRPC Gateway
- Swagger/OpenAPI спецификация генерируется автоматически из proto файлов

## Быстрый старт

### Требования

**Для запуска приложения:**
- Docker и Docker Compose
- Порты должны быть свободны: 8080, 50051, 50052, 5433, 6379, 9092, 2181

**Для разработки (опционально):**
- Go 1.24+
- Инструменты для генерации proto файлов (см. раздел ниже)

### Работа с Proto файлами

**Важно:** Сгенерированные файлы (*.pb.go) уже включены в репозиторий, поэтому вы можете **сразу запускать приложение** без установки дополнительных инструментов.

Установка инструментов для генерации нужна **только если вы изменяете `.proto` файлы**.

#### Установка инструментов для генерации proto (опционально)

<details>
<summary>📦 Развернуть инструкцию по установке</summary>

##### 1. Установите Protocol Buffers Compiler (protoc)

**macOS:**
```bash
brew install protobuf
protoc --version  # Должна быть версия 3.x или выше
```

**Linux:**
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install -y protobuf-compiler
protoc --version

# Или скачайте последнюю версию:
PB_VERSION=25.1
curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v${PB_VERSION}/protoc-${PB_VERSION}-linux-x86_64.zip
unzip protoc-${PB_VERSION}-linux-x86_64.zip -d $HOME/.local
export PATH="$PATH:$HOME/.local/bin"
```

**Windows:**
```powershell
# Через Chocolatey
choco install protoc

# Или скачайте с https://github.com/protocolbuffers/protobuf/releases
```

##### 2. Установите Go плагины для protoc

```bash
# protoc-gen-go (для генерации Go структур)
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# protoc-gen-go-grpc (для генерации gRPC кода)
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# protoc-gen-grpc-gateway (для генерации HTTP gateway)
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest

# protoc-gen-openapiv2 (для генерации Swagger/OpenAPI спецификации)
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

# Проверьте, что плагины в PATH
export PATH="$PATH:$(go env GOPATH)/bin"
```

##### 3. Скачайте Google API definitions

Для генерации proto файлов нужны Google API definitions (используются для HTTP аннотаций):

```bash
# Из корня проекта
git clone --depth 1 https://github.com/googleapis/googleapis.git temp_googleapis
mv temp_googleapis/google ./google
rm -rf temp_googleapis
```

##### 4. Генерация proto файлов

После установки всех инструментов выполните:

```bash
make proto
```

</details>

#### Быстрая команда для установки всего (macOS/Linux)

```bash
# macOS
brew install protobuf && \
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest && \
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest && \
git clone --depth 1 https://github.com/googleapis/googleapis.git temp && \
mv temp/google ./google && rm -rf temp

# Linux (Ubuntu/Debian)
sudo apt install -y protobuf-compiler && \
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest && \
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest && \
git clone --depth 1 https://github.com/googleapis/googleapis.git temp && \
mv temp/google ./google && rm -rf temp
```

**Примечание:** Папка `google/` игнорируется git'ом (см. `.gitignore`), так как содержит ~8000 файлов и нужна только для генерации.

### Запуск

```bash
# Сборка и запуск всех сервисов
docker-compose up --build

# Или в фоновом режиме
docker-compose up --build -d

# Просмотр логов
docker-compose logs -f

# Остановка
docker-compose down
```

### Проверка статуса

```bash
docker-compose ps
```

## API Endpoints

Базовый URL: `http://localhost:8080`

**Примечание:** Все HTTP endpoints работают через **gRPC Gateway**, который автоматически преобразует HTTP/REST запросы в gRPC вызовы к внутренним сервисам. Это позволяет использовать стандартные HTTP методы (GET, POST, PUT, DELETE) при работе с gRPC-сервисами.

### Аутентификация

- `POST /v1/auth/register` - Регистрация пользователя
- `POST /v1/auth/login` - Логин (возвращает JWT токен)

### Задачи (требуют JWT токен)

- `POST /v1/tasks` - Создание задачи
- `GET /v1/tasks` - Получение списка задач
- `PUT /v1/tasks/{id}/complete` - Отметка задачи как выполненной
- `DELETE /v1/tasks/{id}` - Удаление задачи

### Тестирование API

Для тестирования API вы можете использовать Swagger/OpenAPI спецификацию:
- **Файл**: `api_service/swagger/api_service.swagger.json`

Этот файл содержит полное описание всех endpoints, параметров запросов и ответов. Вы можете:
- Импортировать его в Postman, Insomnia или другие API клиенты
- Использовать для генерации клиентского кода
- Просмотреть в Swagger UI

## Структура проекта

```
.
├── api_service/          # API Gateway сервис
│   └── swagger/          # Swagger/OpenAPI спецификация
├── db_service/           # Сервис работы с БД
├── kafka_service/        # Сервис обработки Kafka событий
├── proto/                # Protocol Buffer определения
└── docker-compose.yaml   # Конфигурация Docker Compose
```

## Полезные команды

Проект использует Makefile для упрощения работы. Посмотреть все доступные команды:

```bash
make help
```

### Основные команды

```bash
# Запуск приложения
make up              # Запуск всех сервисов
make up-d            # Запуск в фоновом режиме
make down            # Остановка всех сервисов
make restart         # Перезапуск
make logs            # Просмотр логов
make status          # Статус контейнеров

# Работа с базами данных
make psql            # Подключиться к PostgreSQL
make redis-cli       # Подключиться к Redis
make redis-monitor   # Мониторинг Redis команд в реальном времени
make redis-flush     # Очистить кэш Redis

# Kafka
make kafka-topics    # Список топиков Kafka

# Proto файлы (если изменяли .proto)
make proto           # Регенерация всех proto файлов

# Очистка
make clean           # Остановить и удалить volumes
make clean-all       # Полная очистка (включая образы)
```

### Просмотр логов конкретного сервиса

```bash
docker-compose logs -f api_service
docker-compose logs -f db_service
docker-compose logs -f kafka_service
```

## Переменные окружения

Основные переменные можно изменить в `docker-compose.yaml`:

- `JWT_SECRET_KEY` - секретный ключ для JWT
- `JWT_TOKEN_DURATION` - время жизни токена (в секундах)
- `DB_USER`, `DB_PASSWORD`, `DB_NAME` - параметры БД
- `KAFKA_BROKERS`, `KAFKA_TOPIC` - параметры Kafka

## Troubleshooting

### Сервисы не запускаются

1. Проверьте, что все порты свободны
2. Проверьте логи: `docker-compose logs`
3. Убедитесь, что Docker имеет достаточно ресурсов

### Ошибки подключения к БД

1. Дождитесь, пока PostgreSQL пройдет healthcheck
2. Проверьте переменные окружения
3. Проверьте логи db_service

### Миграции не применены

Миграции автоматически применяются при первом запуске PostgreSQL. Если нужно применить вручную:

```bash
docker-compose exec postgres psql -U docker -d test_db -f /docker-entrypoint-initdb.d/000001_init_schema.up.sql
```

Или используйте команду:
```bash
make migrate-up
```

