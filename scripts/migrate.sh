#!/bin/bash

# Скрипт для применения миграций PostgreSQL
# Использование: ./scripts/migrate.sh [up|down]

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Загрузка переменных окружения
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
else
    echo -e "${RED}Error: .env file not found${NC}"
    exit 1
fi

# Параметры подключения
DB_HOST=${POSTGRES_HOST:-localhost}
DB_PORT=${POSTGRES_PORT:-5432}
DB_USER=${POSTGRES_USER:-postgres}
DB_NAME=${POSTGRES_DB:-geo_alerts}
DB_PASSWORD=${POSTGRES_PASSWORD:-postgres}

# Директория с миграциями
MIGRATIONS_DIR="./migrations"

# Функция для применения миграции
apply_migration() {
    local migration_file=$1
    local direction=$2
    
    echo -e "${YELLOW}Applying migration: ${migration_file}${NC}"
    
    if [ "$direction" == "up" ]; then
        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$migration_file"
    elif [ "$direction" == "down" ]; then
        # Для down миграций ищем соответствующий файл
        local down_file=$(echo "$migration_file" | sed 's/\.up\.sql$/.down.sql/')
        if [ -f "$down_file" ]; then
            PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$down_file"
        else
            echo -e "${RED}Error: Down migration file not found: ${down_file}${NC}"
            exit 1
        fi
    fi
    
    echo -e "${GREEN}Migration applied successfully${NC}"
}

# Проверка наличия psql
if ! command -v psql &> /dev/null; then
    echo -e "${RED}Error: psql command not found. Please install PostgreSQL client.${NC}"
    exit 1
fi

# Проверка директории миграций
if [ ! -d "$MIGRATIONS_DIR" ]; then
    echo -e "${RED}Error: Migrations directory not found: ${MIGRATIONS_DIR}${NC}"
    exit 1
fi

# Определение направления миграции
DIRECTION=${1:-up}

if [ "$DIRECTION" != "up" ] && [ "$DIRECTION" != "down" ]; then
    echo -e "${RED}Error: Invalid direction. Use 'up' or 'down'${NC}"
    exit 1
fi

echo -e "${GREEN}Starting migrations (direction: ${DIRECTION})${NC}"
echo -e "Database: ${DB_NAME}@${DB_HOST}:${DB_PORT}"

# Применение всех миграций
if [ "$DIRECTION" == "up" ]; then
    for migration in $(ls -1 ${MIGRATIONS_DIR}/*.up.sql | sort); do
        apply_migration "$migration" "up"
    done
else
    # Для down применяем в обратном порядке
    for migration in $(ls -1 ${MIGRATIONS_DIR}/*.down.sql | sort -r); do
        apply_migration "$migration" "down"
    done
fi

echo -e "${GREEN}All migrations completed successfully${NC}"

