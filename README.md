# Geo Alert Core - Ядро системы геооповещений

Backend-сервис на Go для системы геооповещений. Сервис интегрируется с новостным порталом (Django) через вебхуки и предоставляет API для управления инцидентами и проверки координат пользователей.

## Архитектура

Проект построен на принципах **Clean Architecture**:

- **Handler** - HTTP слой (Gin framework)
- **Service** - Бизнес-логика
- **Repository** - Слой данных (PostgreSQL)
- **Infrastructure** - Внешние зависимости (Redis, Webhooks)

## Технологии

- **Go 1.24+**
- **PostgreSQL 15** с расширением PostGIS
- **Redis** (кэширование и очередь)
- **Docker & Docker Compose**

## Требования

- Go 1.24 или выше
- Docker и Docker Compose
- PostgreSQL 15 с PostGIS (или используйте Docker)
- Redis (или используйте Docker)
- ngrok (для тестирования вебхуков)

## Быстрый старт

### 1. Клонирование и настройка

```bash
git clone <repository-url>
cd geo-alert-core
```

### 2. Настройка переменных окружения

Создайте файл `.env` на основе `.env.example`:

```bash
cp .env.example .env
```

Отредактируйте `.env`:

```env
# Server
SERVER_PORT=8080

# PostgreSQL
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=geo_alerts
POSTGRES_SSLMODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# API
API_KEY=your-secret-api-key-change-me

# Webhook
WEBHOOK_URL=http://localhost:9090/webhook
WEBHOOK_RETRY_ATTEMPTS=3
WEBHOOK_RETRY_DELAY_SECONDS=5

# Statistics
STATS_TIME_WINDOW_MINUTES=60
```

### 3. Запуск через Docker Compose

```bash
# Запустить все сервисы
docker-compose up -d

# Проверить логи
docker-compose logs -f app
```

Сервис будет доступен на `http://localhost:8080`

### 4. Запуск миграций

#### Вариант 1: Использование скрипта (рекомендуется)

```bash
# Применить миграции
./scripts/migrate.sh up

# Откатить миграции
./scripts/migrate.sh down
```

#### Вариант 2: Использование Makefile

```bash
# Применить миграции
make migrate-up

# Откатить миграции
make migrate-down
```

#### Вариант 3: Вручную

```bash
# Через Docker
docker exec -i geo-alert-postgres psql -U postgres -d geo_alerts < migrations/001_initial.up.sql

# Или локально
psql -h localhost -U postgres -d geo_alerts -f migrations/001_initial.up.sql
```

### 5. Запуск без Docker (локально)

#### Вариант 1: Использование Makefile (рекомендуется)

```bash
# Установить зависимости
make deps

# Запустить сервер
make run

# Или собрать и запустить
make build
./bin/server
```

#### Вариант 2: Напрямую через Go

```bash
# Установить зависимости
go mod download

# Запустить сервер
go run cmd/server/main.go
```

## Полезные команды (Makefile)

```bash
make help          # Показать все доступные команды
make build         # Собрать приложение
make run           # Запустить приложение
make test          # Запустить тесты
make test-coverage # Тесты с покрытием
make migrate-up    # Применить миграции
make migrate-down  # Откатить миграции
make docker-up     # Запустить Docker Compose
make docker-down   # Остановить Docker Compose
make docker-logs   # Показать логи
make fmt           # Форматировать код
make clean         # Очистить скомпилированные файлы
```

## API Endpoints

### Публичные эндпоинты (без API key)

#### Health Check
```bash
GET /api/v1/system/health
```

**Ответ:**
```json
{
  "status": "ok",
  "service": "geo-alert-core"
}
```

#### Проверка координат
```bash
POST /api/v1/location/check
Content-Type: application/json

{
  "user_id": "user123",
  "latitude": 55.7558,
  "longitude": 37.6173
}
```

**Ответ:**
```json
{
  "has_danger": true,
  "incidents": [
    {
      "id": "uuid",
      "title": "Опасная зона",
      "description": "Описание",
      "latitude": 55.7558,
      "longitude": 37.6173,
      "radius": 100.0,
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### Защищенные эндпоинты (требуют API key)

Все запросы должны содержать заголовок:
```
Authorization: Bearer <your-api-key>
```
или
```
X-API-Key: <your-api-key>
```

#### Создание инцидента
```bash
POST /api/v1/incidents
Authorization: Bearer your-api-key
Content-Type: application/json

{
  "title": "Опасная зона",
  "description": "Описание опасности",
  "latitude": 55.7558,
  "longitude": 37.6173,
  "radius": 100.0
}
```

#### Получение всех инцидентов (с пагинацией)
```bash
GET /api/v1/incidents?page=1&page_size=20
Authorization: Bearer your-api-key
```

**Ответ:**
```json
{
  "data": [...],
  "page": 1,
  "page_size": 20
}
```

#### Получение инцидента по ID
```bash
GET /api/v1/incidents/{id}
Authorization: Bearer your-api-key
```

#### Обновление инцидента
```bash
PUT /api/v1/incidents/{id}
Authorization: Bearer your-api-key
Content-Type: application/json

{
  "title": "Обновленное название",
  "is_active": false
}
```

#### Удаление (деактивация) инцидента
```bash
DELETE /api/v1/incidents/{id}
Authorization: Bearer your-api-key
```

#### Статистика по инцидентам
```bash
GET /api/v1/incidents/stats?minutes=60
Authorization: Bearer your-api-key
```

**Ответ:**
```json
{
  "data": [
    {
      "zone_id": "uuid",
      "user_count": 15
    }
  ]
}
```

## Примеры запросов (curl)

### Health Check
```bash
curl http://localhost:8080/api/v1/system/health
```

### Проверка координат
```bash
curl -X POST http://localhost:8080/api/v1/location/check \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "latitude": 55.7558,
    "longitude": 37.6173
  }'
```

### Создание инцидента
```bash
curl -X POST http://localhost:8080/api/v1/incidents \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Опасная зона",
    "description": "Описание",
    "latitude": 55.7558,
    "longitude": 37.6173,
    "radius": 100.0
  }'
```

### Получение всех инцидентов
```bash
curl -X GET "http://localhost:8080/api/v1/incidents?page=1&page_size=20" \
  -H "Authorization: Bearer your-api-key"
```

### Получение инцидента по ID
```bash
curl -X GET http://localhost:8080/api/v1/incidents/{id} \
  -H "Authorization: Bearer your-api-key"
```

### Обновление инцидента
```bash
curl -X PUT http://localhost:8080/api/v1/incidents/{id} \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Новое название",
    "is_active": true
  }'
```

### Удаление инцидента
```bash
curl -X DELETE http://localhost:8080/api/v1/incidents/{id} \
  -H "Authorization: Bearer your-api-key"
```

### Статистика
```bash
curl -X GET "http://localhost:8080/api/v1/incidents/stats?minutes=60" \
  -H "Authorization: Bearer your-api-key"
```

## Настройка ngrok для тестирования вебхуков

### 1. Установка ngrok

```bash
# macOS
brew install ngrok

# Или скачайте с https://ngrok.com/download
```

### 2. Запуск ngrok

```bash
# Запустите локальный сервер для приема вебхуков (например, на порту 9090)
# Затем запустите ngrok
ngrok http 9090
```

Вы получите публичный URL, например: `https://abc123.ngrok.io`

### 3. Настройка WEBHOOK_URL

Обновите `.env`:

```env
WEBHOOK_URL=https://abc123.ngrok.io/webhook
```

Или установите переменную окружения:

```bash
export WEBHOOK_URL=https://abc123.ngrok.io/webhook
```

### 4. Тестирование вебхука

Создайте простой HTTP сервер для приема вебхуков:

```python
# webhook_receiver.py
from http.server import HTTPServer, BaseHTTPRequestHandler
import json

class WebhookHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        content_length = int(self.headers['Content-Length'])
        body = self.rfile.read(content_length)
        data = json.loads(body.decode('utf-8'))
        
        print("Received webhook:")
        print(json.dumps(data, indent=2))
        
        self.send_response(200)
        self.end_headers()
        self.wfile.write(b'OK')

if __name__ == '__main__':
    server = HTTPServer(('localhost', 9090), WebhookHandler)
    print("Webhook receiver listening on http://localhost:9090")
    server.serve_forever()
```

Запустите:
```bash
python3 webhook_receiver.py
```

Теперь при проверке координат с опасными зонами, вебхук будет отправлен на ваш ngrok URL.

## Структура проекта

```
geo-alert-core/
├── cmd/
│   └── server/
│       └── main.go              # Точка входа
├── internal/
│   ├── config/                  # Конфигурация
│   ├── domain/                  # Доменные модели
│   ├── handler/                 # HTTP handlers
│   ├── service/                 # Бизнес-логика
│   ├── repository/              # Слой данных
│   ├── infrastructure/          # Внешние зависимости
│   │   ├── postgres/
│   │   ├── redis/
│   │   └── webhook/
│   └── middleware/              # Middleware (auth)
├── migrations/                  # SQL миграции
│   ├── 001_initial.up.sql
│   └── 001_initial.down.sql
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

## Тестирование

### Запуск unit тестов

```bash
go test ./...
```

### Запуск тестов с покрытием

```bash
go test -cover ./...
```

### Запуск конкретного теста

```bash
go test ./internal/service/... -v
```

## Особенности реализации

### Кэширование

Активные инциденты кэшируются в Redis на 5 минут для ускорения проверки координат.

### Асинхронная отправка вебхуков

Вебхуки отправляются асинхронно в отдельной горутине, чтобы не блокировать ответ клиенту.

### Retry механизм

При неудачной отправке вебхука используется экспоненциальный backoff:
- Попытка 1: сразу
- Попытка 2: через 5 секунд
- Попытка 3: через 10 секунд
- Попытка 4: через 20 секунд

### Геопространственные запросы

Используется PostGIS с GIST индексами для быстрого поиска инцидентов в радиусе от точки.

### Защита от SQL инъекций

Все SQL запросы используют параметризованные запросы (prepared statements).

## Мониторинг

### Health Check

```bash
curl http://localhost:8080/api/v1/system/health
```

### Логи

Логи выводятся в stdout/stderr. При использовании Docker:

```bash
docker-compose logs -f app
```

## Разработка

### Добавление новой миграции

1. Создайте файл `migrations/002_<description>.up.sql`
2. Создайте файл `migrations/002_<description>.down.sql`
3. Примените миграцию вручную или через скрипт

### Форматирование кода

```bash
go fmt ./...
```

### Линтинг

```bash
golangci-lint run
```

## Troubleshooting

### Проблема: "Failed to connect to database"

- Проверьте, что PostgreSQL запущен
- Проверьте настройки в `.env`
- Убедитесь, что PostGIS расширение установлено

### Проблема: "Failed to connect to Redis"

- Проверьте, что Redis запущен
- Проверьте настройки в `.env`

### Проблема: "API_KEY is required"

- Убедитесь, что в `.env` установлен `API_KEY`

### Проблема: Вебхуки не отправляются

- Проверьте `WEBHOOK_URL` в `.env`
- Убедитесь, что ngrok запущен (если используется)
- Проверьте логи приложения


