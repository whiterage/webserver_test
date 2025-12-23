.PHONY: help build run test clean migrate-up migrate-down docker-up docker-down docker-logs

help: ## Показать справку
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Собрать приложение
	@echo "Building application..."
	go build -o bin/server ./cmd/server

run: ## Запустить приложение локально
	@echo "Running application..."
	go run cmd/server/main.go

test: ## Запустить тесты
	@echo "Running tests..."
	go test -v ./...

test-coverage: ## Запустить тесты с покрытием
	@echo "Running tests with coverage..."
	go test -cover ./...

clean: ## Очистить скомпилированные файлы
	@echo "Cleaning..."
	rm -rf bin/
	go clean

migrate-up: ## Применить миграции (up)
	@echo "Applying migrations..."
	./scripts/migrate.sh up

migrate-down: ## Откатить миграции (down)
	@echo "Rolling back migrations..."
	./scripts/migrate.sh down

docker-up: ## Запустить все сервисы через Docker Compose
	@echo "Starting Docker services..."
	docker-compose up -d

docker-down: ## Остановить все сервисы
	@echo "Stopping Docker services..."
	docker-compose down

docker-logs: ## Показать логи Docker
	@echo "Showing Docker logs..."
	docker-compose logs -f app

docker-restart: ## Перезапустить сервисы
	@echo "Restarting Docker services..."
	docker-compose restart

fmt: ## Форматировать код
	@echo "Formatting code..."
	go fmt ./...

lint: ## Запустить линтер (если установлен)
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install it: https://golangci-lint.run/"; \
	fi

deps: ## Установить зависимости
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

