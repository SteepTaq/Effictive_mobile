REST-сервис для агрегации данных об онлайн-подписках пользователей.

## Функциональность

CRUDL-операции над записями о подписках
Подсчет суммарной стоимости подписок за период с фильтрацией
PostgreSQL с миграциями
Логирование запросов и ошибок
Конфигурация через .env
Swagger документация
Docker Compose для запуска
Health check endpoint

## Запуск

```bash
make infra-build
```

Сервис будет доступен на `http://localhost:8080`

## API Endpoints

`GET /health` - проверка здоровья сервиса
`GET /api-docs/` - Swagger документация
`POST /subscriptions` - создание подписки
`GET /subscriptions` - список подписок (с фильтрацией)
`GET /subscriptions/{id}` - получение подписки по ID
`PUT /subscriptions/{id}` - обновление подписки
`DELETE /subscriptions/{id}` - удаление подписки
`GET /subscriptions/summary` - суммарная стоимость за период

## Миграции

Для управления миграциями используйте:

```bash
make migrate-up
make migrate-down
make migrate-reset
```

## Переменные окружения

APP_PORT=8080
APP_ENV=development
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=subscriptions
DB_SSL_MODE=disable
LOG_LEVEL=info

## Makefile команды

```bash
make build         # Собрать приложение
make run           # Собрать и запустить приложение
make clean         # Удалить артефакты сборки
make infra-up      # Запустить инфраструктуру
make infra-down    # Остановить инфраструктуру
make infra-build   # Пересобрать и запустить инфраструктуру
make migrate-up    # Применить миграции
make migrate-down  # Откатить миграции
make migrate-reset # Сбросить миграции
```

## Технологии

Go 1.24.3
PostgreSQL - база данных
Chi Router - HTTP роутер
pgx** - PostgreSQL драйвер
**golang-migrate - миграции
Swagger- документация API
Docker- контейнеризация
